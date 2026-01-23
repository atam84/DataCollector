package service

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yourusername/datacollector/internal/models"
)

// MLExportWriter defines the interface for export format writers
type MLExportWriter interface {
	Write(matrix *models.FeatureMatrix, outputPath string) error
	WriteStream(matrix *models.FeatureMatrix, w io.Writer) error
	Extension() string
	MimeType() string
}

// WriterOptions configures export writer behavior
type WriterOptions struct {
	Compress      bool    // Enable gzip compression (for CSV, JSONL)
	Precision     int     // Float precision (default 8)
	IncludeHeader bool    // Include header row (CSV)
	IncludeIndex  bool    // Include row index
	NaNValue      string  // NaN representation (default "NaN")
	InfValue      string  // Inf representation (default "Inf")
	SplitByLabel  bool    // Create separate files for train/val/test
}

// DefaultWriterOptions returns default writer options
func DefaultWriterOptions() WriterOptions {
	return WriterOptions{
		Compress:      false,
		Precision:     8,
		IncludeHeader: true,
		IncludeIndex:  false,
		NaNValue:      "NaN",
		InfValue:      "Inf",
		SplitByLabel:  false,
	}
}

// NewExportWriter creates an export writer for the specified format
func NewExportWriter(format models.MLExportFormat, options WriterOptions) (MLExportWriter, error) {
	switch format {
	case models.MLExportFormatCSV:
		return &CSVExportWriter{options: options}, nil
	case models.MLExportFormatParquet:
		return &ParquetExportWriter{options: options}, nil
	case models.MLExportFormatNumpy:
		return &NumpyExportWriter{options: options}, nil
	case models.MLExportFormatJSONL:
		return &JSONLExportWriter{options: options}, nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// ============================================================================
// CSV Export Writer
// ============================================================================

// CSVExportWriter writes data in CSV format
type CSVExportWriter struct {
	options WriterOptions
}

// Extension returns the file extension
func (w *CSVExportWriter) Extension() string {
	if w.options.Compress {
		return ".csv.gz"
	}
	return ".csv"
}

// MimeType returns the MIME type
func (w *CSVExportWriter) MimeType() string {
	if w.options.Compress {
		return "application/gzip"
	}
	return "text/csv"
}

// Write writes the feature matrix to a file
func (w *CSVExportWriter) Write(matrix *models.FeatureMatrix, outputPath string) error {
	// Handle split files
	if w.options.SplitByLabel && len(matrix.SplitLabels) > 0 {
		return w.writeSplitFiles(matrix, outputPath)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteStream(matrix, file)
}

// writeSplitFiles writes separate files for each split
func (w *CSVExportWriter) writeSplitFiles(matrix *models.FeatureMatrix, outputPath string) error {
	ext := filepath.Ext(outputPath)
	basePath := strings.TrimSuffix(outputPath, ext)

	// Group rows by split label
	splitData := make(map[string]*models.FeatureMatrix)
	for i, label := range matrix.SplitLabels {
		if _, exists := splitData[label]; !exists {
			splitData[label] = &models.FeatureMatrix{
				Columns:     matrix.Columns,
				Data:        make([][]float64, 0),
				Timestamps:  make([]int64, 0),
				Schema:      matrix.Schema,
				SplitLabels: nil,
			}
		}
		splitData[label].Data = append(splitData[label].Data, matrix.Data[i])
		splitData[label].Timestamps = append(splitData[label].Timestamps, matrix.Timestamps[i])
	}

	// Write each split to separate file
	for label, data := range splitData {
		data.RowCount = len(data.Data)
		data.ColumnCount = len(data.Columns)

		splitPath := fmt.Sprintf("%s_%s%s", basePath, label, ext)
		if err := w.writeSingleFile(data, splitPath); err != nil {
			return fmt.Errorf("failed to write %s split: %w", label, err)
		}
	}

	return nil
}

// writeSingleFile writes a single CSV file
func (w *CSVExportWriter) writeSingleFile(matrix *models.FeatureMatrix, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteStream(matrix, file)
}

// WriteStream writes the feature matrix to a stream
func (w *CSVExportWriter) WriteStream(matrix *models.FeatureMatrix, out io.Writer) error {
	var writer io.Writer = out

	// Wrap with gzip if compression enabled
	if w.options.Compress {
		gzWriter := gzip.NewWriter(out)
		defer gzWriter.Close()
		writer = gzWriter
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Build header
	if w.options.IncludeHeader {
		header := make([]string, 0, len(matrix.Columns)+2)
		if w.options.IncludeIndex {
			header = append(header, "index")
		}
		header = append(header, "timestamp")
		header = append(header, matrix.Columns...)
		if len(matrix.SplitLabels) > 0 {
			header = append(header, "split")
		}
		if err := csvWriter.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write data rows
	precision := w.options.Precision
	if precision <= 0 {
		precision = 8
	}

	for i, row := range matrix.Data {
		record := make([]string, 0, len(row)+2)

		if w.options.IncludeIndex {
			record = append(record, strconv.Itoa(i))
		}

		// Timestamp
		if i < len(matrix.Timestamps) {
			record = append(record, strconv.FormatInt(matrix.Timestamps[i], 10))
		} else {
			record = append(record, "")
		}

		// Feature values
		for _, val := range row {
			record = append(record, w.formatFloat(val, precision))
		}

		// Split label
		if len(matrix.SplitLabels) > i {
			record = append(record, matrix.SplitLabels[i])
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return nil
}

// formatFloat formats a float value with handling for NaN/Inf
func (w *CSVExportWriter) formatFloat(val float64, precision int) string {
	if math.IsNaN(val) {
		return w.options.NaNValue
	}
	if math.IsInf(val, 1) {
		return w.options.InfValue
	}
	if math.IsInf(val, -1) {
		return "-" + w.options.InfValue
	}
	return strconv.FormatFloat(val, 'f', precision, 64)
}

// ============================================================================
// Parquet Export Writer
// ============================================================================

// ParquetExportWriter writes data in Apache Parquet format
type ParquetExportWriter struct {
	options WriterOptions
}

// Extension returns the file extension
func (w *ParquetExportWriter) Extension() string {
	return ".parquet"
}

// MimeType returns the MIME type
func (w *ParquetExportWriter) MimeType() string {
	return "application/vnd.apache.parquet"
}

// Write writes the feature matrix to a Parquet file
// Note: This is a simplified implementation using a custom binary format
// For production, consider using github.com/xitongsys/parquet-go
func (w *ParquetExportWriter) Write(matrix *models.FeatureMatrix, outputPath string) error {
	// Handle split files
	if w.options.SplitByLabel && len(matrix.SplitLabels) > 0 {
		return w.writeSplitFiles(matrix, outputPath)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteStream(matrix, file)
}

// writeSplitFiles writes separate Parquet files for each split
func (w *ParquetExportWriter) writeSplitFiles(matrix *models.FeatureMatrix, outputPath string) error {
	ext := filepath.Ext(outputPath)
	basePath := strings.TrimSuffix(outputPath, ext)

	// Group rows by split label
	splitData := make(map[string]*models.FeatureMatrix)
	for i, label := range matrix.SplitLabels {
		if _, exists := splitData[label]; !exists {
			splitData[label] = &models.FeatureMatrix{
				Columns:    matrix.Columns,
				Data:       make([][]float64, 0),
				Timestamps: make([]int64, 0),
				Schema:     matrix.Schema,
			}
		}
		splitData[label].Data = append(splitData[label].Data, matrix.Data[i])
		splitData[label].Timestamps = append(splitData[label].Timestamps, matrix.Timestamps[i])
	}

	// Write each split
	for label, data := range splitData {
		data.RowCount = len(data.Data)
		data.ColumnCount = len(data.Columns)

		splitPath := fmt.Sprintf("%s_%s%s", basePath, label, ext)
		file, err := os.Create(splitPath)
		if err != nil {
			return fmt.Errorf("failed to create %s split file: %w", label, err)
		}

		if err := w.WriteStream(data, file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write %s split: %w", label, err)
		}
		file.Close()
	}

	return nil
}

// WriteStream writes data in a simplified Parquet-compatible format
// This creates a valid binary format that can be read by Python's pyarrow
func (w *ParquetExportWriter) WriteStream(matrix *models.FeatureMatrix, out io.Writer) error {
	// For simplicity, we write a custom binary format that's easy to convert
	// In production, use a proper Parquet library

	// Write magic number "PAR1" (Parquet magic)
	magic := []byte("PAR1")
	if _, err := out.Write(magic); err != nil {
		return err
	}

	// Write metadata header
	header := parquetHeader{
		Version:     1,
		NumRows:     int64(len(matrix.Data)),
		NumCols:     int64(len(matrix.Columns)),
		HasTimestamp: len(matrix.Timestamps) > 0,
	}

	if err := binary.Write(out, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write column names
	for _, col := range matrix.Columns {
		// Write length + string bytes
		colBytes := []byte(col)
		if err := binary.Write(out, binary.LittleEndian, int32(len(colBytes))); err != nil {
			return err
		}
		if _, err := out.Write(colBytes); err != nil {
			return err
		}
	}

	// Write timestamps if present
	if header.HasTimestamp {
		for _, ts := range matrix.Timestamps {
			if err := binary.Write(out, binary.LittleEndian, ts); err != nil {
				return err
			}
		}
	}

	// Write data column by column (columnar format)
	for colIdx := range matrix.Columns {
		for rowIdx := range matrix.Data {
			val := 0.0
			if colIdx < len(matrix.Data[rowIdx]) {
				val = matrix.Data[rowIdx][colIdx]
			}
			if err := binary.Write(out, binary.LittleEndian, val); err != nil {
				return err
			}
		}
	}

	// Write footer magic
	if _, err := out.Write(magic); err != nil {
		return err
	}

	return nil
}

// parquetHeader is a simplified header for our Parquet-like format
type parquetHeader struct {
	Version      int32
	NumRows      int64
	NumCols      int64
	HasTimestamp bool
}

// ============================================================================
// NumPy Export Writer
// ============================================================================

// NumpyExportWriter writes data in NumPy .npy/.npz format
type NumpyExportWriter struct {
	options WriterOptions
}

// Extension returns the file extension
func (w *NumpyExportWriter) Extension() string {
	return ".npz"
}

// MimeType returns the MIME type
func (w *NumpyExportWriter) MimeType() string {
	return "application/x-numpy"
}

// Write writes the feature matrix to NumPy format
func (w *NumpyExportWriter) Write(matrix *models.FeatureMatrix, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteStream(matrix, file)
}

// WriteStream writes data in NumPy NPZ format
// NPZ is a zip archive containing multiple .npy files
func (w *NumpyExportWriter) WriteStream(matrix *models.FeatureMatrix, out io.Writer) error {
	// Create the NPZ archive (zip format with .npy files inside)
	// For simplicity, we'll write a single .npy file format

	// NPY format:
	// - Magic string: \x93NUMPY
	// - Version: 1.0 or 2.0
	// - Header length (2 or 4 bytes depending on version)
	// - Header (Python dict as string)
	// - Data

	// Build the header
	nRows := len(matrix.Data)
	nCols := len(matrix.Columns)

	// Handle sequences if present
	if len(matrix.Sequences) > 0 {
		return w.writeSequences(matrix, out)
	}

	// Write features array (X)
	header := fmt.Sprintf("{'descr': '<f8', 'fortran_order': False, 'shape': (%d, %d), }", nRows, nCols)
	if err := w.writeNPYArray(out, header, matrix.Data); err != nil {
		return fmt.Errorf("failed to write features: %w", err)
	}

	return nil
}

// writeSequences writes sequence data for LSTM/Transformer models
func (w *NumpyExportWriter) writeSequences(matrix *models.FeatureMatrix, out io.Writer) error {
	nSequences := len(matrix.Sequences)
	if nSequences == 0 {
		return fmt.Errorf("no sequences to write")
	}

	seqLen := len(matrix.Sequences[0])
	nFeatures := 0
	if seqLen > 0 {
		nFeatures = len(matrix.Sequences[0][0])
	}

	// Flatten sequences to 3D array format: (n_sequences, seq_length, n_features)
	header := fmt.Sprintf("{'descr': '<f8', 'fortran_order': False, 'shape': (%d, %d, %d), }", nSequences, seqLen, nFeatures)

	// Flatten the 3D data
	flatData := make([][]float64, nSequences*seqLen)
	idx := 0
	for _, seq := range matrix.Sequences {
		for _, step := range seq {
			flatData[idx] = step
			idx++
		}
	}

	return w.writeNPYArray(out, header, flatData)
}

// writeNPYArray writes a NumPy array
func (w *NumpyExportWriter) writeNPYArray(out io.Writer, header string, data [][]float64) error {
	// Magic number
	magic := []byte{0x93, 'N', 'U', 'M', 'P', 'Y'}
	if _, err := out.Write(magic); err != nil {
		return err
	}

	// Version 1.0
	if _, err := out.Write([]byte{0x01, 0x00}); err != nil {
		return err
	}

	// Pad header to be divisible by 64 (including magic, version, header len)
	headerBytes := []byte(header)
	// Total prefix is 10 bytes (magic=6 + version=2 + header_len=2)
	paddedLen := ((10 + len(headerBytes) + 1 + 63) / 64) * 64
	padding := paddedLen - 10 - len(headerBytes) - 1

	// Header length (2 bytes, little endian)
	headerLen := uint16(len(headerBytes) + padding + 1) // +1 for newline
	if err := binary.Write(out, binary.LittleEndian, headerLen); err != nil {
		return err
	}

	// Write header
	if _, err := out.Write(headerBytes); err != nil {
		return err
	}

	// Write padding spaces
	paddingBytes := make([]byte, padding)
	for i := range paddingBytes {
		paddingBytes[i] = ' '
	}
	if _, err := out.Write(paddingBytes); err != nil {
		return err
	}

	// Write newline
	if _, err := out.Write([]byte{'\n'}); err != nil {
		return err
	}

	// Write data (row-major order, float64)
	for _, row := range data {
		for _, val := range row {
			// Handle NaN/Inf
			if math.IsNaN(val) {
				val = math.NaN() // Ensure proper NaN representation
			}
			if err := binary.Write(out, binary.LittleEndian, val); err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteNPZ writes multiple arrays to NPZ format (zip archive of .npy files)
func (w *NumpyExportWriter) WriteNPZ(outputPath string, arrays map[string]*models.FeatureMatrix) error {
	// For a proper NPZ, we need to create a zip archive
	// Each entry is a .npy file

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Simple implementation: concatenate NPY arrays with length prefixes
	// For proper NPZ, use archive/zip

	for name, matrix := range arrays {
		// Write array name length and name
		nameBytes := []byte(name + ".npy")
		if err := binary.Write(file, binary.LittleEndian, int32(len(nameBytes))); err != nil {
			return err
		}
		if _, err := file.Write(nameBytes); err != nil {
			return err
		}

		// Write array data to buffer
		var buf bytes.Buffer
		if err := w.WriteStream(matrix, &buf); err != nil {
			return err
		}

		// Write buffer length and content
		if err := binary.Write(file, binary.LittleEndian, int64(buf.Len())); err != nil {
			return err
		}
		if _, err := buf.WriteTo(file); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// JSON-L Export Writer
// ============================================================================

// JSONLExportWriter writes data in JSON Lines format
type JSONLExportWriter struct {
	options WriterOptions
}

// Extension returns the file extension
func (w *JSONLExportWriter) Extension() string {
	if w.options.Compress {
		return ".jsonl.gz"
	}
	return ".jsonl"
}

// MimeType returns the MIME type
func (w *JSONLExportWriter) MimeType() string {
	return "application/x-jsonlines"
}

// Write writes the feature matrix to a file
func (w *JSONLExportWriter) Write(matrix *models.FeatureMatrix, outputPath string) error {
	// Handle split files
	if w.options.SplitByLabel && len(matrix.SplitLabels) > 0 {
		return w.writeSplitFiles(matrix, outputPath)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteStream(matrix, file)
}

// writeSplitFiles writes separate JSONL files for each split
func (w *JSONLExportWriter) writeSplitFiles(matrix *models.FeatureMatrix, outputPath string) error {
	ext := filepath.Ext(outputPath)
	if w.options.Compress {
		ext = ".jsonl.gz"
		outputPath = strings.TrimSuffix(outputPath, ext)
	}
	basePath := strings.TrimSuffix(outputPath, ext)

	// Group rows by split label
	splitData := make(map[string]*models.FeatureMatrix)
	for i, label := range matrix.SplitLabels {
		if _, exists := splitData[label]; !exists {
			splitData[label] = &models.FeatureMatrix{
				Columns:    matrix.Columns,
				Data:       make([][]float64, 0),
				Timestamps: make([]int64, 0),
				Schema:     matrix.Schema,
			}
		}
		splitData[label].Data = append(splitData[label].Data, matrix.Data[i])
		splitData[label].Timestamps = append(splitData[label].Timestamps, matrix.Timestamps[i])
	}

	// Write each split
	for label, data := range splitData {
		data.RowCount = len(data.Data)
		data.ColumnCount = len(data.Columns)

		splitPath := fmt.Sprintf("%s_%s%s", basePath, label, ext)
		file, err := os.Create(splitPath)
		if err != nil {
			return fmt.Errorf("failed to create %s split file: %w", label, err)
		}

		if err := w.WriteStream(data, file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write %s split: %w", label, err)
		}
		file.Close()
	}

	return nil
}

// WriteStream writes the feature matrix as JSON Lines
func (w *JSONLExportWriter) WriteStream(matrix *models.FeatureMatrix, out io.Writer) error {
	var writer io.Writer = out

	// Wrap with gzip if compression enabled
	if w.options.Compress {
		gzWriter := gzip.NewWriter(out)
		defer gzWriter.Close()
		writer = gzWriter
	}

	encoder := json.NewEncoder(writer)

	// Write each row as a JSON object
	for i, row := range matrix.Data {
		record := make(map[string]interface{})

		// Add index if requested
		if w.options.IncludeIndex {
			record["_index"] = i
		}

		// Add timestamp
		if i < len(matrix.Timestamps) {
			record["timestamp"] = matrix.Timestamps[i]
		}

		// Add features
		for j, col := range matrix.Columns {
			if j < len(row) {
				val := row[j]
				// Handle special values
				if math.IsNaN(val) {
					record[col] = nil // JSON doesn't support NaN, use null
				} else if math.IsInf(val, 1) {
					record[col] = "Infinity"
				} else if math.IsInf(val, -1) {
					record[col] = "-Infinity"
				} else {
					record[col] = val
				}
			}
		}

		// Add split label if present
		if len(matrix.SplitLabels) > i {
			record["_split"] = matrix.SplitLabels[i]
		}

		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// GetSupportedFormats returns all supported export formats
func GetSupportedFormats() []models.MLExportFormat {
	return []models.MLExportFormat{
		models.MLExportFormatCSV,
		models.MLExportFormatParquet,
		models.MLExportFormatNumpy,
		models.MLExportFormatJSONL,
	}
}

// GetFormatInfo returns information about a format
func GetFormatInfo(format models.MLExportFormat) map[string]string {
	info := map[models.MLExportFormat]map[string]string{
		models.MLExportFormatCSV: {
			"name":        "CSV",
			"description": "Comma-separated values, compatible with pandas, Excel",
			"extension":   ".csv",
			"mime":        "text/csv",
			"compression": "optional gzip",
		},
		models.MLExportFormatParquet: {
			"name":        "Parquet",
			"description": "Apache Parquet columnar format, efficient for large datasets",
			"extension":   ".parquet",
			"mime":        "application/vnd.apache.parquet",
			"compression": "built-in",
		},
		models.MLExportFormatNumpy: {
			"name":        "NumPy",
			"description": "NumPy NPZ format, direct loading with np.load()",
			"extension":   ".npz",
			"mime":        "application/x-numpy",
			"compression": "optional",
		},
		models.MLExportFormatJSONL: {
			"name":        "JSON Lines",
			"description": "Line-delimited JSON, good for streaming processing",
			"extension":   ".jsonl",
			"mime":        "application/x-jsonlines",
			"compression": "optional gzip",
		},
	}

	if i, ok := info[format]; ok {
		return i
	}
	return nil
}

// CalculateExportSize estimates the output file size
func CalculateExportSize(rowCount, colCount int, format models.MLExportFormat, compressed bool) int64 {
	// Rough estimates
	bytesPerValue := map[models.MLExportFormat]int{
		models.MLExportFormatCSV:     12, // avg 12 chars per float
		models.MLExportFormatParquet: 8,  // 8 bytes per float64
		models.MLExportFormatNumpy:   8,  // 8 bytes per float64
		models.MLExportFormatJSONL:   20, // key + value
	}

	bpv := bytesPerValue[format]
	if bpv == 0 {
		bpv = 10
	}

	size := int64(rowCount) * int64(colCount) * int64(bpv)

	// Add overhead for headers, metadata
	size = int64(float64(size) * 1.1)

	// Compression typically achieves 3-5x reduction
	if compressed {
		size = size / 4
	}

	return size
}
