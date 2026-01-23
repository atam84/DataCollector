package service

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MLExportService handles ML data export operations
type MLExportService struct {
	ohlcvRepo    *repository.OHLCVRepository
	jobRepo      *repository.JobRepository
	exportRepo   *repository.MLExportRepository
	featureEngine *MLFeatureEngine

	// Background job management
	activeJobs   map[string]context.CancelFunc
	activeJobsMu sync.RWMutex

	// Configuration
	exportDir string
}

// NewMLExportService creates a new ML export service
func NewMLExportService(
	ohlcvRepo *repository.OHLCVRepository,
	jobRepo *repository.JobRepository,
	exportRepo *repository.MLExportRepository,
) *MLExportService {
	// Default export directory
	exportDir := os.Getenv("ML_EXPORT_DIR")
	if exportDir == "" {
		exportDir = "/tmp/ml_exports"
	}

	// Ensure directory exists
	os.MkdirAll(exportDir, 0755)

	return &MLExportService{
		ohlcvRepo:     ohlcvRepo,
		jobRepo:       jobRepo,
		exportRepo:    exportRepo,
		featureEngine: NewMLFeatureEngine(),
		activeJobs:    make(map[string]context.CancelFunc),
		exportDir:     exportDir,
	}
}

// StartExport starts a background export job
func (s *MLExportService) StartExport(ctx context.Context, config models.MLExportConfig, jobIDs []primitive.ObjectID) (*models.MLExportJob, error) {
	// Validate inputs
	if len(jobIDs) == 0 {
		return nil, fmt.Errorf("no job IDs provided")
	}

	// Create export job record
	exportJob := &models.MLExportJob{
		Status:       models.MLExportStatusPending,
		JobIDs:       jobIDs,
		Config:       config,
		Progress:     0,
		CurrentPhase: "initializing",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := s.exportRepo.CreateExportJob(ctx, exportJob); err != nil {
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	// Start background processing
	go s.processExportJob(exportJob.ID.Hex())

	return exportJob, nil
}

// processExportJob handles the background export processing
func (s *MLExportService) processExportJob(exportJobID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Track active job
	s.activeJobsMu.Lock()
	s.activeJobs[exportJobID] = cancel
	s.activeJobsMu.Unlock()

	defer func() {
		s.activeJobsMu.Lock()
		delete(s.activeJobs, exportJobID)
		s.activeJobsMu.Unlock()
	}()

	// Get export job
	objID, _ := primitive.ObjectIDFromHex(exportJobID)
	exportJob, err := s.exportRepo.FindExportJobByID(ctx, objID)
	if err != nil {
		s.failExportJob(ctx, objID, fmt.Sprintf("failed to find export job: %v", err))
		return
	}

	// Update status to running
	now := time.Now()
	exportJob.Status = models.MLExportStatusRunning
	exportJob.StartedAt = &now
	exportJob.CurrentPhase = "loading"
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Load candle data from all source jobs
	allCandles, sourceInfos, err := s.loadCandleData(ctx, exportJob)
	if err != nil {
		s.failExportJob(ctx, objID, fmt.Sprintf("failed to load data: %v", err))
		return
	}

	exportJob.TotalRecords = int64(len(allCandles))
	exportJob.Progress = 10
	exportJob.CurrentPhase = "features"
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Generate features
	matrix, err := s.featureEngine.GenerateFeatures(allCandles, exportJob.Config.Features)
	if err != nil {
		s.failExportJob(ctx, objID, fmt.Sprintf("failed to generate features: %v", err))
		return
	}

	exportJob.Progress = 30
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Generate targets
	if err := s.featureEngine.GenerateTargets(matrix, allCandles, exportJob.Config.Target); err != nil {
		s.failExportJob(ctx, objID, fmt.Sprintf("failed to generate targets: %v", err))
		return
	}

	exportJob.Progress = 40
	exportJob.CurrentPhase = "preprocessing"
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Apply preprocessing
	normParams, err := s.applyPreprocessing(matrix, exportJob.Config.Preprocessing)
	if err != nil {
		s.failExportJob(ctx, objID, fmt.Sprintf("failed to apply preprocessing: %v", err))
		return
	}

	exportJob.Progress = 50
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Apply split if enabled
	var splitInfo *models.SplitInfo
	if exportJob.Config.Split.Enabled {
		splitInfo = s.applySplit(matrix, exportJob.Config.Split)
	}

	exportJob.Progress = 60
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Generate sequences if enabled
	var seqInfo *models.SequenceInfo
	if exportJob.Config.Sequence.Enabled {
		seqInfo = s.generateSequences(matrix, exportJob.Config.Sequence)
	}

	exportJob.Progress = 70
	exportJob.CurrentPhase = "writing"
	s.exportRepo.UpdateExportJob(ctx, exportJob)

	// Write output file
	outputPath, fileSize, err := s.writeOutput(matrix, exportJob)
	if err != nil {
		s.failExportJob(ctx, objID, fmt.Sprintf("failed to write output: %v", err))
		return
	}

	// Build metadata
	exportJob.Metadata = s.buildMetadata(matrix, allCandles, sourceInfos, normParams, splitInfo, seqInfo)
	exportJob.OutputPath = outputPath
	exportJob.FileSizeBytes = fileSize
	exportJob.FeatureCount = matrix.ColumnCount
	exportJob.ColumnNames = matrix.Columns
	exportJob.RowCount = int64(matrix.RowCount)
	exportJob.ProcessedRecords = int64(len(allCandles))

	// Complete the job
	completedAt := time.Now()
	expiresAt := completedAt.Add(24 * time.Hour) // Files expire after 24 hours
	exportJob.Status = models.MLExportStatusCompleted
	exportJob.CompletedAt = &completedAt
	exportJob.ExpiresAt = &expiresAt
	exportJob.Progress = 100
	exportJob.CurrentPhase = "completed"

	s.exportRepo.UpdateExportJob(ctx, exportJob)
}

// loadCandleData loads candle data from source jobs
func (s *MLExportService) loadCandleData(ctx context.Context, exportJob *models.MLExportJob) ([]models.Candle, []models.SourceJobInfo, error) {
	var allCandles []models.Candle
	var sourceInfos []models.SourceJobInfo

	for _, jobID := range exportJob.JobIDs {
		// Get job info
		job, err := s.jobRepo.FindByID(ctx, jobID.Hex())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find job %s: %w", jobID.Hex(), err)
		}

		// Load OHLCV data using job's exchange, symbol, and timeframe
		doc, err := s.ohlcvRepo.FindByJob(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load candles for job %s: %w", jobID.Hex(), err)
		}

		if doc == nil || len(doc.Candles) == 0 {
			continue
		}

		candles := doc.Candles

		// Sort candles
		sort.Slice(candles, func(i, j int) bool {
			return candles[i].Timestamp < candles[j].Timestamp
		})

		// Track source info
		sourceInfo := models.SourceJobInfo{
			JobID:      jobID,
			ExchangeID: job.ConnectorExchangeID,
			Symbol:     job.Symbol,
			Timeframe:  job.Timeframe,
			BarCount:   int64(len(candles)),
			StartTime:  time.UnixMilli(candles[0].Timestamp),
			EndTime:    time.UnixMilli(candles[len(candles)-1].Timestamp),
		}
		sourceInfos = append(sourceInfos, sourceInfo)

		allCandles = append(allCandles, candles...)
	}

	// Sort all candles by timestamp
	sort.Slice(allCandles, func(i, j int) bool {
		return allCandles[i].Timestamp < allCandles[j].Timestamp
	})

	return allCandles, sourceInfos, nil
}

// applyPreprocessing applies preprocessing steps to the feature matrix
func (s *MLExportService) applyPreprocessing(matrix *models.FeatureMatrix, config models.PreprocessConfig) (map[string]models.NormParams, error) {
	normParams := make(map[string]models.NormParams)

	// Handle NaN values first
	s.handleNaN(matrix, config.NaNHandling)

	// Handle Inf values
	s.handleInf(matrix, config.InfHandling)

	// Clip outliers if enabled
	if config.ClipOutliers && config.OutlierStdDev > 0 {
		s.clipOutliers(matrix, config.OutlierStdDev)
	}

	// Apply normalization
	if config.Normalization != models.NormalizationNone {
		normParams = s.normalize(matrix, config.Normalization)
	}

	// Remove NaN rows if requested (after all preprocessing)
	if config.RemoveNaNRows {
		s.removeNaNRows(matrix)
	}

	return normParams, nil
}

// handleNaN handles NaN values based on strategy
func (s *MLExportService) handleNaN(matrix *models.FeatureMatrix, handling models.NaNHandlingType) {
	for colIdx := range matrix.Columns {
		switch handling {
		case models.NaNHandlingZero:
			for rowIdx := range matrix.Data {
				if math.IsNaN(matrix.Data[rowIdx][colIdx]) {
					matrix.Data[rowIdx][colIdx] = 0
				}
			}

		case models.NaNHandlingForwardFill:
			lastValid := 0.0
			for rowIdx := range matrix.Data {
				if math.IsNaN(matrix.Data[rowIdx][colIdx]) {
					matrix.Data[rowIdx][colIdx] = lastValid
				} else {
					lastValid = matrix.Data[rowIdx][colIdx]
				}
			}

		case models.NaNHandlingBackwardFill:
			nextValid := 0.0
			for rowIdx := len(matrix.Data) - 1; rowIdx >= 0; rowIdx-- {
				if math.IsNaN(matrix.Data[rowIdx][colIdx]) {
					matrix.Data[rowIdx][colIdx] = nextValid
				} else {
					nextValid = matrix.Data[rowIdx][colIdx]
				}
			}

		case models.NaNHandlingInterpolate:
			// Linear interpolation
			for rowIdx := range matrix.Data {
				if math.IsNaN(matrix.Data[rowIdx][colIdx]) {
					// Find prev and next valid values
					prevIdx, nextIdx := -1, -1
					for i := rowIdx - 1; i >= 0; i-- {
						if !math.IsNaN(matrix.Data[i][colIdx]) {
							prevIdx = i
							break
						}
					}
					for i := rowIdx + 1; i < len(matrix.Data); i++ {
						if !math.IsNaN(matrix.Data[i][colIdx]) {
							nextIdx = i
							break
						}
					}

					if prevIdx >= 0 && nextIdx >= 0 {
						// Interpolate
						prevVal := matrix.Data[prevIdx][colIdx]
						nextVal := matrix.Data[nextIdx][colIdx]
						ratio := float64(rowIdx-prevIdx) / float64(nextIdx-prevIdx)
						matrix.Data[rowIdx][colIdx] = prevVal + ratio*(nextVal-prevVal)
					} else if prevIdx >= 0 {
						matrix.Data[rowIdx][colIdx] = matrix.Data[prevIdx][colIdx]
					} else if nextIdx >= 0 {
						matrix.Data[rowIdx][colIdx] = matrix.Data[nextIdx][colIdx]
					} else {
						matrix.Data[rowIdx][colIdx] = 0
					}
				}
			}
		}
	}
}

// handleInf handles infinite values
func (s *MLExportService) handleInf(matrix *models.FeatureMatrix, handling string) {
	if handling == "" {
		handling = "replace_nan"
	}

	for colIdx := range matrix.Columns {
		for rowIdx := range matrix.Data {
			val := matrix.Data[rowIdx][colIdx]
			if math.IsInf(val, 1) || math.IsInf(val, -1) {
				switch handling {
				case "replace_nan":
					matrix.Data[rowIdx][colIdx] = math.NaN()
				case "clip":
					if math.IsInf(val, 1) {
						matrix.Data[rowIdx][colIdx] = math.MaxFloat64 / 2
					} else {
						matrix.Data[rowIdx][colIdx] = -math.MaxFloat64 / 2
					}
				case "drop":
					matrix.Data[rowIdx][colIdx] = math.NaN()
				}
			}
		}
	}
}

// clipOutliers clips values beyond N standard deviations
func (s *MLExportService) clipOutliers(matrix *models.FeatureMatrix, nStdDev float64) {
	for colIdx := range matrix.Columns {
		// Calculate mean and std for column
		sum := 0.0
		count := 0
		for _, row := range matrix.Data {
			if !math.IsNaN(row[colIdx]) {
				sum += row[colIdx]
				count++
			}
		}
		if count == 0 {
			continue
		}
		mean := sum / float64(count)

		sumSq := 0.0
		for _, row := range matrix.Data {
			if !math.IsNaN(row[colIdx]) {
				diff := row[colIdx] - mean
				sumSq += diff * diff
			}
		}
		std := math.Sqrt(sumSq / float64(count))

		// Clip values
		lower := mean - nStdDev*std
		upper := mean + nStdDev*std

		for rowIdx := range matrix.Data {
			if !math.IsNaN(matrix.Data[rowIdx][colIdx]) {
				if matrix.Data[rowIdx][colIdx] < lower {
					matrix.Data[rowIdx][colIdx] = lower
				} else if matrix.Data[rowIdx][colIdx] > upper {
					matrix.Data[rowIdx][colIdx] = upper
				}
			}
		}
	}
}

// normalize applies normalization to the feature matrix
func (s *MLExportService) normalize(matrix *models.FeatureMatrix, method models.NormalizationType) map[string]models.NormParams {
	params := make(map[string]models.NormParams)

	for colIdx, colName := range matrix.Columns {
		// Skip target columns
		if len(colName) > 7 && colName[:7] == "target_" {
			continue
		}

		// Calculate stats
		values := make([]float64, 0, len(matrix.Data))
		for _, row := range matrix.Data {
			if !math.IsNaN(row[colIdx]) {
				values = append(values, row[colIdx])
			}
		}
		if len(values) == 0 {
			continue
		}

		sort.Float64s(values)

		var normParam models.NormParams
		normParam.Method = string(method)

		switch method {
		case models.NormalizationMinMax:
			min := values[0]
			max := values[len(values)-1]
			normParam.Min = min
			normParam.Max = max

			range_ := max - min
			if range_ > 0 {
				for rowIdx := range matrix.Data {
					if !math.IsNaN(matrix.Data[rowIdx][colIdx]) {
						matrix.Data[rowIdx][colIdx] = (matrix.Data[rowIdx][colIdx] - min) / range_
					}
				}
			}

		case models.NormalizationZScore:
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			mean := sum / float64(len(values))

			sumSq := 0.0
			for _, v := range values {
				diff := v - mean
				sumSq += diff * diff
			}
			std := math.Sqrt(sumSq / float64(len(values)))

			normParam.Mean = mean
			normParam.Std = std
			normParam.Min = values[0]
			normParam.Max = values[len(values)-1]

			if std > 0 {
				for rowIdx := range matrix.Data {
					if !math.IsNaN(matrix.Data[rowIdx][colIdx]) {
						matrix.Data[rowIdx][colIdx] = (matrix.Data[rowIdx][colIdx] - mean) / std
					}
				}
			}

		case models.NormalizationRobust:
			// Use median and IQR
			median := values[len(values)/2]
			q1 := values[len(values)/4]
			q3 := values[3*len(values)/4]
			iqr := q3 - q1

			normParam.Median = median
			normParam.IQR = iqr
			normParam.Min = values[0]
			normParam.Max = values[len(values)-1]

			if iqr > 0 {
				for rowIdx := range matrix.Data {
					if !math.IsNaN(matrix.Data[rowIdx][colIdx]) {
						matrix.Data[rowIdx][colIdx] = (matrix.Data[rowIdx][colIdx] - median) / iqr
					}
				}
			}
		}

		params[colName] = normParam
	}

	return params
}

// removeNaNRows removes rows with any NaN values
func (s *MLExportService) removeNaNRows(matrix *models.FeatureMatrix) {
	newData := make([][]float64, 0, len(matrix.Data))
	newTimestamps := make([]int64, 0, len(matrix.Timestamps))
	var newSplitLabels []string
	if len(matrix.SplitLabels) > 0 {
		newSplitLabels = make([]string, 0, len(matrix.SplitLabels))
	}

	for i, row := range matrix.Data {
		hasNaN := false
		for _, val := range row {
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
		}

		if !hasNaN {
			newData = append(newData, row)
			if i < len(matrix.Timestamps) {
				newTimestamps = append(newTimestamps, matrix.Timestamps[i])
			}
			if len(matrix.SplitLabels) > i {
				newSplitLabels = append(newSplitLabels, matrix.SplitLabels[i])
			}
		}
	}

	matrix.Data = newData
	matrix.Timestamps = newTimestamps
	matrix.SplitLabels = newSplitLabels
	matrix.RowCount = len(newData)
}

// applySplit applies train/validation/test split
func (s *MLExportService) applySplit(matrix *models.FeatureMatrix, config models.SplitConfig) *models.SplitInfo {
	n := len(matrix.Data)
	if n == 0 {
		return nil
	}

	// Calculate split indices
	trainEnd := int(float64(n) * config.TrainRatio)
	valEnd := trainEnd + int(float64(n)*config.ValidationRatio)

	// Initialize split labels
	matrix.SplitLabels = make([]string, n)

	splitInfo := &models.SplitInfo{}

	if config.TimeBased {
		// Time-based split (chronological)
		for i := 0; i < n; i++ {
			if i < trainEnd {
				matrix.SplitLabels[i] = "train"
			} else if i < valEnd {
				matrix.SplitLabels[i] = "validation"
			} else {
				matrix.SplitLabels[i] = "test"
			}
		}

		// Record split info
		splitInfo.TrainRows = int64(trainEnd)
		splitInfo.ValRows = int64(valEnd - trainEnd)
		splitInfo.TestRows = int64(n - valEnd)

		if trainEnd > 0 {
			splitInfo.TrainStart = time.UnixMilli(matrix.Timestamps[0])
			splitInfo.TrainEnd = time.UnixMilli(matrix.Timestamps[trainEnd-1])
		}
		if valEnd > trainEnd {
			splitInfo.ValStart = time.UnixMilli(matrix.Timestamps[trainEnd])
			splitInfo.ValEnd = time.UnixMilli(matrix.Timestamps[valEnd-1])
		}
		if n > valEnd {
			splitInfo.TestStart = time.UnixMilli(matrix.Timestamps[valEnd])
			splitInfo.TestEnd = time.UnixMilli(matrix.Timestamps[n-1])
		}

	} else if config.Shuffle {
		// Random shuffle split
		indices := make([]int, n)
		for i := range indices {
			indices[i] = i
		}

		// Simple shuffle (Fisher-Yates)
		for i := n - 1; i > 0; i-- {
			j := int(time.Now().UnixNano()) % (i + 1)
			indices[i], indices[j] = indices[j], indices[i]
		}

		for i, idx := range indices {
			if i < trainEnd {
				matrix.SplitLabels[idx] = "train"
			} else if i < valEnd {
				matrix.SplitLabels[idx] = "validation"
			} else {
				matrix.SplitLabels[idx] = "test"
			}
		}

		// Count for info
		trainCount, valCount, testCount := int64(0), int64(0), int64(0)
		for _, label := range matrix.SplitLabels {
			switch label {
			case "train":
				trainCount++
			case "validation":
				valCount++
			case "test":
				testCount++
			}
		}
		splitInfo.TrainRows = trainCount
		splitInfo.ValRows = valCount
		splitInfo.TestRows = testCount
	}

	return splitInfo
}

// generateSequences creates sequences for RNN/LSTM/Transformer models
func (s *MLExportService) generateSequences(matrix *models.FeatureMatrix, config models.SequenceConfig) *models.SequenceInfo {
	if config.Length <= 0 || len(matrix.Data) < config.Length {
		return nil
	}

	stride := config.Stride
	if stride <= 0 {
		stride = 1
	}

	sequences := make([][][]float64, 0)

	for i := 0; i <= len(matrix.Data)-config.Length; i += stride {
		seq := make([][]float64, config.Length)
		for j := 0; j < config.Length; j++ {
			seq[j] = make([]float64, len(matrix.Data[i+j]))
			copy(seq[j], matrix.Data[i+j])
		}
		sequences = append(sequences, seq)
	}

	matrix.Sequences = sequences

	// Calculate sequence counts per split
	seqInfo := &models.SequenceInfo{
		Length:         config.Length,
		Stride:         stride,
		TotalSequences: int64(len(sequences)),
	}

	// If split labels exist, count sequences per split
	if len(matrix.SplitLabels) > 0 {
		for i := 0; i <= len(matrix.Data)-config.Length; i += stride {
			// Sequence belongs to the split of its last element
			lastIdx := i + config.Length - 1
			if lastIdx < len(matrix.SplitLabels) {
				switch matrix.SplitLabels[lastIdx] {
				case "train":
					seqInfo.TrainSequences++
				case "validation":
					seqInfo.ValSequences++
				case "test":
					seqInfo.TestSequences++
				}
			}
		}
	}

	return seqInfo
}

// writeOutput writes the feature matrix to the output file
func (s *MLExportService) writeOutput(matrix *models.FeatureMatrix, exportJob *models.MLExportJob) (string, int64, error) {
	// Create writer
	options := DefaultWriterOptions()
	options.SplitByLabel = exportJob.Config.Split.Enabled

	writer, err := NewExportWriter(exportJob.Config.Format, options)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create writer: %w", err)
	}

	// Generate output filename
	filename := fmt.Sprintf("ml_export_%s_%s%s",
		exportJob.ID.Hex(),
		time.Now().Format("20060102_150405"),
		writer.Extension(),
	)
	outputPath := filepath.Join(s.exportDir, filename)

	// Write file
	if err := writer.Write(matrix, outputPath); err != nil {
		return "", 0, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return outputPath, 0, nil
	}

	return outputPath, fileInfo.Size(), nil
}

// buildMetadata creates export metadata
func (s *MLExportService) buildMetadata(
	matrix *models.FeatureMatrix,
	candles []models.Candle,
	sourceInfos []models.SourceJobInfo,
	normParams map[string]models.NormParams,
	splitInfo *models.SplitInfo,
	seqInfo *models.SequenceInfo,
) models.MLExportMetadata {
	metadata := models.MLExportMetadata{
		Version:       "1.0",
		ExportedAt:    time.Now(),
		SourceJobs:    sourceInfos,
		FeatureSchema: matrix.Schema,
	}

	if len(candles) > 0 {
		// Extract unique values
		symbols := make(map[string]bool)
		timeframes := make(map[string]bool)
		exchanges := make(map[string]bool)

		for _, info := range sourceInfos {
			symbols[info.Symbol] = true
			timeframes[info.Timeframe] = true
			exchanges[info.ExchangeID] = true
		}

		metadata.DataRange = models.DataRange{
			StartTime:  time.UnixMilli(candles[0].Timestamp),
			EndTime:    time.UnixMilli(candles[len(candles)-1].Timestamp),
			TotalBars:  int64(len(candles)),
			Symbols:    mapKeys(symbols),
			Timeframes: mapKeys(timeframes),
			Exchanges:  mapKeys(exchanges),
		}
	}

	if len(normParams) > 0 {
		metadata.NormalizationParams = normParams
	}

	if splitInfo != nil {
		metadata.SplitInfo = splitInfo
	}

	if seqInfo != nil {
		metadata.SequenceInfo = seqInfo
	}

	return metadata
}

// failExportJob marks an export job as failed
func (s *MLExportService) failExportJob(ctx context.Context, jobID primitive.ObjectID, errMsg string) {
	exportJob, err := s.exportRepo.FindExportJobByID(ctx, jobID)
	if err != nil {
		return
	}

	exportJob.Status = models.MLExportStatusFailed
	exportJob.LastError = errMsg
	exportJob.Errors = append(exportJob.Errors, errMsg)
	now := time.Now()
	exportJob.CompletedAt = &now
	exportJob.UpdatedAt = now

	s.exportRepo.UpdateExportJob(ctx, exportJob)
}

// GetExportJob retrieves an export job by ID
func (s *MLExportService) GetExportJob(ctx context.Context, id string) (*models.MLExportJob, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid export job ID: %w", err)
	}
	return s.exportRepo.FindExportJobByID(ctx, objID)
}

// ListExportJobs lists export jobs with pagination
func (s *MLExportService) ListExportJobs(ctx context.Context, limit, offset int64) ([]*models.MLExportJob, int64, error) {
	return s.exportRepo.FindExportJobs(ctx, limit, offset)
}

// CancelExportJob cancels a running export job
func (s *MLExportService) CancelExportJob(ctx context.Context, id string) error {
	s.activeJobsMu.RLock()
	cancel, exists := s.activeJobs[id]
	s.activeJobsMu.RUnlock()

	if exists {
		cancel()
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid export job ID: %w", err)
	}

	exportJob, err := s.exportRepo.FindExportJobByID(ctx, objID)
	if err != nil {
		return err
	}

	exportJob.Status = models.MLExportStatusCancelled
	now := time.Now()
	exportJob.CompletedAt = &now
	exportJob.UpdatedAt = now

	return s.exportRepo.UpdateExportJob(ctx, exportJob)
}

// DeleteExportJob deletes an export job and its output file
func (s *MLExportService) DeleteExportJob(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid export job ID: %w", err)
	}

	// Get job to find output file
	exportJob, err := s.exportRepo.FindExportJobByID(ctx, objID)
	if err != nil {
		return err
	}

	// Delete output file if exists
	if exportJob.OutputPath != "" {
		os.Remove(exportJob.OutputPath)
	}

	// Delete from database
	return s.exportRepo.DeleteExportJob(ctx, objID)
}

// StreamExport performs a streaming export (for smaller datasets)
func (s *MLExportService) StreamExport(ctx context.Context, config models.MLExportConfig, jobID primitive.ObjectID, writer *os.File) error {
	// Get job info first
	job, err := s.jobRepo.FindByID(ctx, jobID.Hex())
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	// Load candle data using job's exchange, symbol, and timeframe
	doc, err := s.ohlcvRepo.FindByJob(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
	if err != nil {
		return fmt.Errorf("failed to load candles: %w", err)
	}

	if doc == nil || len(doc.Candles) == 0 {
		return fmt.Errorf("no candle data found for job")
	}

	candles := doc.Candles

	// Generate features
	matrix, err := s.featureEngine.GenerateFeatures(candles, config.Features)
	if err != nil {
		return fmt.Errorf("failed to generate features: %w", err)
	}

	// Generate targets
	if err := s.featureEngine.GenerateTargets(matrix, candles, config.Target); err != nil {
		return fmt.Errorf("failed to generate targets: %w", err)
	}

	// Apply preprocessing
	_, err = s.applyPreprocessing(matrix, config.Preprocessing)
	if err != nil {
		return fmt.Errorf("failed to apply preprocessing: %w", err)
	}

	// Apply split if enabled
	if config.Split.Enabled {
		s.applySplit(matrix, config.Split)
	}

	// Generate sequences if enabled
	if config.Sequence.Enabled {
		s.generateSequences(matrix, config.Sequence)
	}

	// Create export writer
	options := DefaultWriterOptions()
	exportWriter, err := NewExportWriter(config.Format, options)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}

	// Write to stream
	return exportWriter.WriteStream(matrix, writer)
}

// GetPresets returns all available export presets
func (s *MLExportService) GetPresets(ctx context.Context) ([]models.MLExportConfig, error) {
	return models.GetBuiltinPresets(), nil
}

// GetConfig retrieves an export configuration by ID
func (s *MLExportService) GetConfig(ctx context.Context, id string) (*models.MLExportConfig, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}
	return s.exportRepo.FindConfigByID(ctx, objID)
}

// CreateConfig creates a new export configuration
func (s *MLExportService) CreateConfig(ctx context.Context, config *models.MLExportConfig) error {
	return s.exportRepo.CreateConfig(ctx, config)
}

// UpdateConfig updates an export configuration
func (s *MLExportService) UpdateConfig(ctx context.Context, config *models.MLExportConfig) error {
	return s.exportRepo.UpdateConfig(ctx, config)
}

// DeleteConfig deletes an export configuration
func (s *MLExportService) DeleteConfig(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}
	return s.exportRepo.DeleteConfig(ctx, objID)
}

// ListConfigs lists saved export configurations
func (s *MLExportService) ListConfigs(ctx context.Context) ([]*models.MLExportConfig, error) {
	return s.exportRepo.FindConfigs(ctx)
}

// CleanupExpiredExports removes expired export files
func (s *MLExportService) CleanupExpiredExports(ctx context.Context) error {
	jobs, _, err := s.exportRepo.FindExportJobs(ctx, 1000, 0)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, job := range jobs {
		if job.ExpiresAt != nil && job.ExpiresAt.Before(now) {
			// Remove file
			if job.OutputPath != "" {
				os.Remove(job.OutputPath)
			}
			// Delete job record
			s.exportRepo.DeleteExportJob(ctx, job.ID)
		}
	}

	return nil
}

// Helper function to get map keys
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
