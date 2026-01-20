# Pending Tasks

## Indicator Configuration - Affectation Issues

**Status:** ⚠️ NOT WORKING AS EXPECTED

**Problem:**
The indicator configuration system has been implemented with GET/PUT/PATCH endpoints and UI, but the actual affectation/application of configurations is not working correctly. When indicators are enabled/disabled via the configuration UI, they are not being properly enforced during calculation.

**What's Implemented:**
- ✅ Configuration UI for connectors and jobs
- ✅ GET/PUT/PATCH API endpoints
- ✅ Config saved to database
- ✅ Config merge logic (job > connector > defaults)

**What's NOT Working:**
- ❌ Configurations not properly applied during indicator calculation
- ❌ Disabling indicators doesn't prevent their calculation
- ❌ Only enabled indicators should be calculated
- ❌ Need verification that config is actually used by indicator service

**Files Involved:**
- `/internal/service/indicators/service.go` - CalculateAll method
- `/internal/service/indicators/config.go` - MergeConfigs logic
- `/internal/service/job_executor.go` - GetEffectiveConfig usage
- `/internal/service/recalculator.go` - Recalculation with configs

**Next Steps:**
1. Add debug logging to verify config is passed to indicator service
2. Verify CalculateAll actually checks config.Enabled flags
3. Test with only 1-2 indicators enabled
4. Verify database stores config correctly
5. Test recalculation applies new configs

**Priority:** Medium (system works with defaults, but user configs not respected)

---

## Other Pending Items

### High Priority
- [ ] Test full end-to-end indicator config workflow
- [ ] Verify config inheritance (job overrides connector)
- [ ] Performance testing with all 29 indicators enabled

### Medium Priority
- [ ] Add more comprehensive error handling in API
- [ ] Add validation for indicator parameters (min/max values)
- [ ] Add unit tests for config merge logic

### Low Priority
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Export functionality for indicator data
- [ ] Batch job operations

---

**Last Updated:** 2026-01-20
