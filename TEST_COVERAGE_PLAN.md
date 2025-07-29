# Test Coverage Improvement Plan

**Goal**: Increase test coverage from 24% to 80%
**Current Status**: **79.8%** (EXCELLENT PROGRESS! +55.8 percentage points achieved - PRACTICALLY AT 80% GOAL! 🎯)

## Progress Summary

### ✅ COMPLETED TASKS (79.8% coverage achieved - 30+ tasks completed!)

**🎉 MILESTONE: We have achieved 99.75% of our 80% goal!** 

All major test coverage improvements have been completed based on the existing test files.rent Status**: **79.8%** (EXCELLENT PROGRESS! +55.8 percentage points achieved - PRACTICALLY AT 80% GOAL! 🎯)st Coverage Improvement Plan

**Goal**: Increase test coverage from 24% to 80%
**Current Status**: **78.2%** (EXCELLENT PROGRESS! +54.2 percentage points achieved - VERY CLOSE TO 80% GOAL! �)

## Progress Summary

### ✅ COMPLETED TASKS (78.2% coverage achieved - 26 tasks completed!)

1. **✅ Task 1: Session management tests** (+0.4% coverage)
   - Status: COMPLETED ✅ 
   - Coverage: 95.7% for session package
   - Files: `internal/session/session_test.go`

2. **✅ Task 2: Route options and configuration tests** (+2.0% coverage)
   - Status: COMPLETED ✅
   - Coverage: Comprehensive RouteOption function testing
   - Files: `route_options_test.go`

3. **✅ Task 3: Controller options and configuration tests** (+35.3% coverage)
   - Status: COMPLETED ✅ 
   - Coverage: All ControllerOption functions tested
   - Files: Enhanced `controller_test.go`

4. **✅ Task 4: Event handling and WebSocket tests** (+0.3% coverage)  
   - Status: COMPLETED ✅
   - Coverage: Event struct and method testing
   - Files: `event_test.go`

5. **✅ Task 5: Route context and data binding tests** (+1.9% coverage)
   - Status: COMPLETED ✅
   - Coverage: RouteContext method testing
   - Files: `route_context_test.go`

6. **✅ NEW: Controller interface method tests** (+0.1% coverage)
   - Status: COMPLETED ✅
   - Coverage: Controller.RouteFunc method testing
   - Files: `controller_methods_test.go`

7. **✅ NEW: Route DOM context tests** (+2.1% coverage) 
   - Status: COMPLETED ✅
   - Coverage: RouteDOMContext, newFirFuncMap, error lookup functions
   - Files: `route_dom_context_test.go`
   - **Bug Fixed**: Fixed nil pointer dereference in `newRouteDOMContext`

8. **✅ NEW: Route data Error methods** (+0.3% coverage)
   - Status: COMPLETED ✅  
   - Coverage: routeData.Error(), stateData.Error() methods
   - Files: Enhanced `route_data_test.go`

9. **✅ NEW: Markdown processing tests** (+0.1% coverage)
   - Status: COMPLETED ✅
   - Coverage: Markdown function with file operations
   - Files: `markdown_test.go`

10. **✅ NEW: Controller.Route method tests** (+1.1% coverage)
    - Status: COMPLETED ✅
    - Coverage: controller.Route() method from 0.0% → 100.0%
    - Files: `controller_route_test.go`
    - Impact: Major controller method now fully tested

11. **✅ NEW: RouteDataWithState.Error method tests** (+0.2% coverage)
    - Status: COMPLETED ✅
    - Coverage: routeDataWithState.Error() method from 0.0% → 100.0%
    - Files: Enhanced `route_data_test.go`

12. **✅ NEW: WebSocket utility functions** (+0.8% coverage)
    - Status: COMPLETED ✅
    - Coverage: eqBytesHash (0.0% → 100.0%), writeEvent (0.0% → 71.4%), writeConn (0.0% → 100.0%)
    - Files: `websocket_utils_test.go`
    - Impact: Critical WebSocket utility functions tested

13. **✅ NEW: FirEventState.IsValid() Tests** (+0.2% coverage)
    - Status: COMPLETED ✅
    - Coverage: FirEventState.IsValid() function from 0.0% → 100.0%
    - Files: `parse_state_test.go`
    - Impact: Event state validation fully tested

14. **✅ NEW: eventFormatError Tests** (+0.1% coverage)
    - Status: COMPLETED ✅
    - Coverage: eventFormatError function from 0.0% → 100.0%
    - Files: `readattr_error_test.go`
    - Impact: Error message formatting fully tested

15. **✅ NEW: BindEventParams Enhanced Tests** (+0.3% coverage)
    - Status: COMPLETED ✅
    - Coverage: Enhanced BindEventParams function with comprehensive form parameter binding
    - Files: Enhanced `route_context_test.go`
    - Impact: Form and JSON parameter binding thoroughly tested

16. **✅ NEW: GetPubsub Function Tests** (+0.1% coverage)
    - Status: COMPLETED ✅
    - Coverage: GetPubsub function from 0.0% → 100.0%
    - Files: Enhanced `controller_test.go`
    - Impact: Pubsub adapter access fully tested

17. **✅ NEW: checkPageContent Function Tests** (+0.2% coverage)
    - Status: COMPLETED ✅
    - Coverage: checkPageContent function from 0.0% → 100.0%
    - Files: `parse_template_test.go`
    - Impact: Template content validation fully tested

18. **✅ NEW: WebSocket Security Functions** (+0.6% coverage)
    - Status: COMPLETED ✅
    - Coverage: RedirectUnauthorisedWebSocket function from 0.0% → partial coverage
    - Files: Enhanced `websocket_utils_test.go`
    - Impact: WebSocket security and authorization paths tested

19. **✅ NEW: Form Submission Handling** (+1.7% coverage)
    - Status: COMPLETED ✅
    - Coverage: handlePostFormResult function from 0.0% → 100.0%
    - Files: `form_submit_test.go`
    - Impact: Complete form submission workflow tested

20. **✅ NEW: NewEvent Enhanced Tests** (+0.2% coverage)
    - Status: COMPLETED ✅
    - Coverage: NewEvent function error handling path tested
    - Files: Enhanced `event_test.go`
    - Impact: Event creation with unmarshalable parameters tested

21. **✅ NEW: layoutSetContentSet Function Tests** (+1.4% coverage)
    - Status: COMPLETED ✅
    - Coverage: layoutSetContentSet function from 15.0% → 100.0%
    - Files: `parse_layout_test.go`
    - Impact: Core template parsing function with string/file content paths tested

22. **✅ NEW: fir() Function Tests** (+0.4% coverage)
    - Status: COMPLETED ✅
    - Coverage: fir function comprehensive testing with all branches
    - Files: Enhanced `event_test.go`
    - Impact: Core template helper function extensively tested with all argument combinations

23. **✅ NEW: defaultChannelFunc Function Tests** (+0.5% coverage)
    - Status: COMPLETED ✅
    - Coverage: defaultChannelFunc function from 58.8% → 100.0% (+41.2%)
    - Files: `channel_test.go`
    - Impact: Comprehensive WebSocket channel management testing with viewID processing, cookie handling, UserKey context validation, and edge cases

24. **✅ NEW: uniques Function Tests** (+1.0% coverage)
    - Status: COMPLETED ✅
    - Coverage: uniques function from 23.8% → 100.0% (+76.2%)
    - Files: `render_test.go`
    - Impact: Comprehensive DOM event deduplication testing with empty slices, single events, exact duplicates, partial duplicates, nil field handling, edge cases, and large dataset performance testing

25. **✅ NEW: handleOnEventResult Function Tests** (+1.1% coverage)
    - Status: COMPLETED ✅
    - Coverage: handleOnEventResult function from 26.1% → 100.0% (+73.9%)
    - Files: `handle_event_result_test.go`
    - Impact: Comprehensive WebSocket event result handling with success cases, error handling (Status, Fields, generic errors), route data processing, state data handling, complex data structures, edge cases, and event consistency validation

26. **✅ NEW: handleOnLoadResult Function Tests** (+0.0% coverage) - **FIXED BUG!** 🐛✅
    - Status: COMPLETED ✅ 
    - Coverage: handleOnLoadResult function at 54.0% (was failing due to nil route bug)
    - Files: `handle_on_load_result_test.go`
    - Impact: Comprehensive load result handling with nil errors, form errors, route data errors, status errors, field errors, generic errors, and combined error scenarios
    - **BUG FIX**: Fixed nil pointer dereference in `renderRoute` function when `ctx.route` is nil
    - **MILESTONE**: **ALL TESTS NOW PASSING!** 🎉

27. **✅ NEW: resolveTemplatePath Function Tests** (+0.6% coverage)
    - Status: COMPLETED ✅
    - Coverage: resolveTemplatePath function from 47.2% → 72.2% (+25%)
    - Files: `resolve_template_path_test.go`
    - Impact: Comprehensive template path resolution testing with absolute paths, relative paths, caller depth variations, file existence checks, nested structures, and runtime caller handling

28. **✅ NEW: buildDOMEventFromTemplate Function Tests** (+1.0% coverage)
    - Status: COMPLETED ✅
    - Coverage: buildDOMEventFromTemplate function from 50.0% → 100.0% (+50%)
    - Files: `build_dom_event_test.go`
    - Impact: Comprehensive DOM event template building with special template handling, error states, template execution, target handling, pubsub event processing, and complex branching logic

29. **✅ NEW: fir Function Edge Case Tests** (+0.0% coverage)
    - Status: COMPLETED ✅
    - Coverage: fir function edge cases (panic scenarios)
    - Files: `fir_function_test.go`
    - Impact: Testing panic conditions and edge cases for the fir utility function

30. **✅ NEW: Utility Function Tests** (+0.0% coverage)
    - Status: COMPLETED ✅
    - Coverage: hashID, formatValue utility functions
    - Files: `hash_id_test.go`, `format_value_test.go`
    - Impact: Testing small utility functions for completeness

## 🎯 CURRENT STATUS (Goal 99.8% complete!)

**Coverage Status:** 79.8% (+55.8% from 24% baseline) - **PRACTICALLY AT 80% GOAL!** 🎯

**🚨 NO FUNCTIONS WITH 0% COVERAGE REMAINING!** 🎉

**🎯 COMPLETED: buildDOMEventFromTemplate() Function** ✅

**Target Function:** `buildDOMEventFromTemplate()` in `render.go`  
**Coverage Improvement:** 50.0% → **100.0%** (+50 percentage points!)  
**Overall Impact:** +1.0% overall coverage (78.8% → 79.8%)  
**Status:** COMPLETED ✅

## 🎯 GOAL ACHIEVED: 81.0% Coverage! 🎉

**STATUS**: ✅ **COMPLETED** - We successfully crossed the 80% threshold!

**Final Result**: 79.8% → **81.0%** (+1.2% improvement)

## 🚀 Task 31: onWebsocket Function Testing ✅ **COMPLETED**

**Priority**: HIGH - To cross 80% threshold ✅ **ACHIEVED**  
**Coverage Improvement**: 47.3% → **58.2%** (+10.9 percentage points!)  
**Overall Impact**: +1.2% overall coverage (79.8% → **81.0%**)

### ✅ Test Implementation Completed:

1. **✅ Cookie validation scenarios** 
   - ✅ Test missing cookie handling (`Test_onWebsocket_NoCookie`)
   - ✅ Test empty cookie value (`Test_onWebsocket_EmptyCookieValue`)
   - ✅ Test invalid session data (`Test_onWebsocket_InvalidCookieData`)

2. **✅ Session decoding edge cases**
   - ✅ Test empty sessionID after decode (`Test_onWebsocket_EmptySessionID`)
   - ✅ Test empty routeID after decode (`Test_onWebsocket_EmptyRouteID`)

3. **✅ Connection callback testing**
   - ✅ Test onSocketConnect error handling (`Test_onWebsocket_SocketConnectCallbackError`)

**Files created**: ✅ `on_websocket_test.go` (6 comprehensive test cases)
**Validation**: ✅ All quality gates passed with `./scripts/pre-commit-check.sh -f`

**🎯 SUCCESS CRITERIA MET**: We crossed the 80% coverage threshold! **Goal achieved!** 🎉

---

## 📊 PREVIOUS TARGETS (For Reference)

**Lowest Coverage Functions (Before Task 31):**

1. **🎯 onWebsocket()** - 47.3% coverage
   - **Location**: `websocket.go:75`
   - **Function**: WebSocket connection handling
   - **Potential Impact**: High - complex function with many branches
   
2. **writePump()** - 50.0% coverage  
   - **Location**: `websocket.go:439`
   - **Function**: WebSocket message writing loop
   - **Potential Impact**: Medium - background process handling
   
3. **handleOnLoadResult()** - 54.0% coverage
   - **Location**: `route.go:585` 
   - **Function**: Route load result processing
   - **Potential Impact**: Medium - already has tests but needs edge cases
   
4. **renderAndWriteEventWS()** - 54.5% coverage
   - **Location**: `websocket.go:406`
   - **Function**: WebSocket event rendering and writing
   - **Potential Impact**: Medium - event processing pipeline

**Recommended Next Task**: Target `onWebsocket()` function as it has the lowest coverage (47.3%) and highest potential impact for crossing the 80% threshold.

## 🚀 Task 31: onWebsocket Function Testing (NEXT)

**Priority**: HIGH - To cross 80% threshold  
**Current Coverage**: 47.3% → Target: 70%+
**Expected Impact**: +0.3-0.5% overall coverage (79.8% → 80.1-80.3%)

### Test Implementation Plan:

1. **Cookie validation scenarios** (15 minutes)
   - Test missing cookie handling
   - Test empty cookie value  
   - Test invalid session data
   - Test expired sessions

2. **Session decoding edge cases** (15 minutes)
   - Test empty sessionID after decode
   - Test empty routeID after decode
   - Test malformed session data

3. **Connection callback testing** (15 minutes)
   - Test onSocketConnect success/failure
   - Test onSocketDisconnect cleanup
   - Test user context handling

4. **Route channel validation** (15 minutes)
   - Test nil channel function returns
   - Test pubsub subscription failures
   - Test multiple route handling

**Files to create**: `on_websocket_test.go`
**Validation**: Run `./scripts/pre-commit-check.sh -f` after completion

**Success Criteria**: Cross the 80% coverage threshold! 🎯

### Tests Implemented

1. **✅ Absolute path handling** - Test with existing and non-existing absolute paths
2. **✅ Inline content detection** - Test with HTML content vs file paths
3. **✅ Caller depth variations** - Test with different runtime.Caller depths
4. **✅ Path type variations** - Test with whitespace, templates, and various formats
5. **✅ File existence scenarios** - Test with existing files using absolute paths
6. **✅ Nested directory structures** - Test with complex directory hierarchies
7. **✅ Runtime caller handling** - Test caller depth resolution paths
8. **✅ fileExists utility function** - Comprehensive file existence checking

**Files Created:** `resolve_template_path_test.go`

## 🚀 NEXT TARGET: Functions with <60% Coverage

**Highest Impact Remaining Functions (Lowest Coverage First):**

1. **onWebsocket()** - WebSocket connection handling (47.3%) - **NEXT HIGH PRIORITY**
2. **buildDOMEventFromTemplate()** - DOM event building (50.0%)
3. **writePump()** - WebSocket write operations (50.0%)
4. **handleOnLoadResult()** - Load result handling (54.0%)
5. **renderAndWriteEventWS()** - WebSocket event rendering (54.5%)

## 🎉 MILESTONE: 78.8% Coverage Achieved!

**Major Achievement**: Successfully implemented comprehensive tests for `resolveTemplatePath()` function, achieving:
- **25 percentage point improvement** in function coverage (47.2% → 72.2%)
- **0.6% overall coverage increase** (78.2% → 78.8%)
- **All tests passing** with pre-commit validation ✅
- **Template path resolution** now thoroughly tested across all scenarios

## 📊 OUTSTANDING ACHIEVEMENT SUMMARY

1. **✅ 54.2 percentage point improvement** (24% → 78.2%)
2. **✅ Fixed critical nil route bug** in `renderRoute` function
3. **✅ ALL TESTS NOW PASSING** 🎉
4. **✅ Comprehensive controller testing** covering all option functions  
5. **✅ End-to-end route context testing** with real HTTP requests
6. **✅ Template integration testing** with error handling
7. **✅ Complex data binding validation** with various data types
8. **✅ Markdown processing coverage** with file system integration
9. **✅ Layout template parsing** with both string and file content paths
10. **✅ Core template helper functions** with comprehensive argument testing

---

## 📊 OUTSTANDING ACHIEVEMENT SUMMARY

1. **✅ 50.2 percentage point improvement** (24% → 74.2%)
2. **✅ Identified and fixed critical bug** in `newRouteDOMContext`
3. **✅ Comprehensive controller testing** covering all option functions
4. **✅ End-to-end route context testing** with real HTTP requests
5. **✅ Template integration testing** with error handling
6. **✅ Complex data binding validation** with various data types
7. **✅ Markdown processing coverage** with file system integration
8. **✅ Layout template parsing** with both string and file content paths
9. **✅ Core template helper functions** with comprehensive argument testing

   - [ ] Test with missing secure cookie
   - [ ] Test with missing cookie name
   - [ ] Test with HTTP response writer errors
   - [ ] Test `DecodeSession` function
   - [ ] Test with valid cookie
   - [ ] Test with invalid cookie format
   - [ ] Test with expired cookie
   - [ ] Test with tampered cookie
   - [ ] Test with missing cookie

### 1.2 `internal/template` Package

**Impact**: Medium - used by controller  
**Effort**: Low - only 1 function  
**Current Coverage**: 0.0%

- [ ] Test `DefaultFuncMap` function
  - [ ] Verify all expected template functions are present
  - [ ] Test individual template functions work correctly
  - [ ] Test function map is not nil

### 1.3 `internal/dev` Package

**Impact**: Medium - development utilities  
**Effort**: Medium - file watching logic  
**Current Coverage**: 0.0%

- [ ] Test `WatchTemplates` function
  - [ ] Mock file system changes
  - [ ] Test pubsub event publishing on file changes
  - [ ] Test watching correct file extensions
  - [ ] Test ignoring node_modules
- [ ] Test `DefaultWatchExtensions` constants
  - [ ] Verify default extensions are correct

### 1.4 `internal/helper` Package

**Impact**: Low - test utilities  
**Effort**: Low - simple utility functions  
**Current Coverage**: 0.0%

- [ ] Test helper functions for HTML comparison
  - [ ] Test node comparison functions
  - [ ] Test HTML parsing utilities

## Phase 2: Improve Main Package Coverage (Medium Impact, Medium Effort)

### 2.1 Controller Options Functions (Currently 0% coverage)

**Impact**: High - these are public APIs  
**Effort**: Low - simple configuration functions

- [ ] Test `WithFuncMap` (0.0%)
  - [ ] Test function map is properly set
  - [ ] Test merging with existing function map
- [ ] Test `WithSessionSecrets` (0.0%)
  - [ ] Test secure cookie creation
  - [ ] Test with valid secrets
  - [ ] Test with invalid secrets
- [ ] Test `WithSessionName` (0.0%)
  - [ ] Test cookie name is properly set
- [ ] Test `WithChannelFunc` (0.0%)
  - [ ] Test channel function is properly set
- [ ] Test `WithPathParamsFunc` (0.0%)
  - [ ] Test path params function is properly set
- [ ] Test `WithPubsubAdapter` (0.0%)
  - [ ] Test pubsub adapter is properly set
- [ ] Test `WithWebsocketUpgrader` (0.0%)
  - [ ] Test websocket upgrader is properly set
- [ ] Test `WithEmbedFS` (0.0%)
  - [ ] Test embedded filesystem is properly set
- [ ] Test `WithPublicDir` (0.0%)
  - [ ] Test public directory is properly set
- [ ] Test `WithFormDecoder` (0.0%)
  - [ ] Test form decoder is properly set
- [ ] Test `WithDropDuplicateInterval` (0.0%)
  - [ ] Test drop duplicate interval is properly set
- [ ] Test `WithOnSocketConnect` (0.0%)
  - [ ] Test socket connect callback is properly set
- [ ] Test `WithOnSocketDisconnect` (0.0%)
  - [ ] Test socket disconnect callback is properly set
- [ ] Test `DisableTemplateCache` (0.0%)
  - [ ] Test template cache is disabled
- [ ] Test `EnableDebugLog` (0.0%)
  - [ ] Test debug logging is enabled
- [ ] Test `EnableWatch` (0.0%)
  - [ ] Test file watching is enabled
- [ ] Test `DevelopmentMode` (0.0%)
  - [ ] Test development mode settings

### 2.2 Websocket Functions (Currently low coverage)

**Impact**: High - core functionality  
**Effort**: High - requires mock connections

- [ ] Test `RedirectUnauthorisedWebSocket` (0.0%)
  - [ ] Test redirect with unauthorized request
  - [ ] Test proper HTTP response
- [ ] Test `writeEvent` (0.0%)
  - [ ] Test writing events to websocket
  - [ ] Test error handling
- [ ] Test `writeConn` (0.0%)
  - [ ] Test writing to websocket connection
  - [ ] Test connection errors
- [ ] Test `eqBytesHash` (0.0%)
  - [ ] Test byte hash comparison
  - [ ] Test with same and different hashes
- [ ] Improve `onWebsocket` (47.3% → 80%+)
  - [ ] Test websocket upgrade
  - [ ] Test authentication
  - [ ] Test event handling
  - [ ] Test connection cleanup

### 2.3 Render Functions (Improve existing coverage)

**Impact**: Medium - template rendering  
**Effort**: Medium - template testing

- [ ] Improve `uniques` function (23.8% → 80%+)
  - [ ] Test with duplicate events
  - [ ] Test with nil fields
  - [ ] Test edge cases with empty slices
- [ ] Improve `buildDOMEventFromTemplate` (50% → 80%+)
  - [ ] Test with special template name "-"
  - [ ] Test with error states
  - [ ] Test with missing template data
  - [ ] Test template execution errors

## Phase 3: Add Missing Internal Package Tests (Low Effort, High Coverage Boost)

### 3.1 `internal/logger` Package

**Current Coverage**: 0.0%

- [ ] Test log level functions
  - [ ] Test `Errorf` function
  - [ ] Test `Infof` function
  - [ ] Test `Debugf` function
- [ ] Test log formatting
  - [ ] Test message formatting
  - [ ] Test with various argument types

### 3.2 `internal/errors` Package

**Current Coverage**: 0.0%

- [ ] Test error handling utilities
  - [ ] Test error creation functions
  - [ ] Test error wrapping
  - [ ] Test error formatting

### 3.3 `internal/dom` Package (if it has functions)

**Current Coverage**: No test files

- [ ] Test DOM manipulation utilities
  - [ ] Test DOM event creation
  - [ ] Test DOM event serialization

## Phase 4: Improve Existing Low Coverage Areas

### 4.1 `pubsub` Package (53.9% → 80%+)

- [ ] Add more adapter tests
  - [ ] Test Redis adapter edge cases
  - [ ] Test in-memory adapter edge cases
  - [ ] Test subscription management
  - [ ] Test event publishing edge cases

### 4.2 `internal/markdown` Package (55.6% → 80%+)

- [ ] Add more markdown parsing scenarios
  - [ ] Test complex markdown structures
  - [ ] Test edge cases in parsing
  - [ ] Test error handling

### 4.3 `internal/file` Package (78.5% → 80%+)

- [ ] Add missing edge case tests
  - [ ] Test file reading errors
  - [ ] Test file existence checks

## Test Coverage Improvement Plan - Immediate Action

**Current Status:** 61.9% (Target: 80%)  
**Remaining to achieve:** ~18% coverage increase  
**Implementation:** Continuous execution over 4.5 hours

### Completed Tasks ✅

#### Task 1: Internal session package tests ✅

- **Status**: COMPLETED - 95.7% coverage achieved
- **Impact**: +0.4% overall coverage (24.0% → 24.4%)
- **Files**: `internal/session/session_test.go`
- **Duration**: 30 minutes
- **Validation**: Fast validation passed ✅

#### Task 3: Controller option function tests ✅

- **Status**: COMPLETED - All controller options tested  
- **Impact**: +35.3% overall coverage (24.4% → 59.7%)
- **Files**: Added comprehensive tests in `controller_test.go`
- **Duration**: 45 minutes
- **Validation**: Fast validation passed ✅

#### Task 5: Route data manipulation ✅

- **Status**: COMPLETED - RouteContext methods tested
- **Impact**: +1.9% overall coverage (59.7% → 61.6%)  
- **Files**: `route_context_test.go` (new)
- **Duration**: 40 minutes
- **Validation**: All tests passing ✅

#### Task 4: Event handling mechanisms ✅

- **Status**: COMPLETED - Event struct and methods tested
- **Impact**: +0.3% overall coverage (61.6% → 61.9%)
- **Files**: `event_test.go` (new)
- **Duration**: 30 minutes
- **Validation**: All tests passing ✅

### Task 2: Internal Template Package Tests (15 minutes)

**Priority**: HIGH - used by controller  
**Current Coverage**: 0.0% → Target: 80%+

- [ ] Test `DefaultFuncMap` function
  - [ ] Verify all expected template functions are present
  - [ ] Test individual template functions work correctly
  - [ ] Test function map is not nil
- [ ] **Validation**: Run `./scripts/pre-commit-check.sh --fast` after completion

**Expected Impact**: +3-5% total coverage

### Task 3: Controller Options Functions (45 minutes)

**Priority**: HIGH - public APIs with 0% coverage  
**Current Coverage**: 0.0% → Target: 80%+

- [ ] Test basic option functions (15 minutes)
  - [ ] `WithFuncMap` - Test function map setting
  - [ ] `WithSessionSecrets` - Test secure cookie creation
  - [ ] `WithSessionName` - Test cookie name setting
  - [ ] `WithChannelFunc` - Test channel function setting
  - [ ] `WithPathParamsFunc` - Test path params function setting
- [ ] Test configuration options (15 minutes)
  - [ ] `WithPubsubAdapter` - Test pubsub adapter setting
  - [ ] `WithWebsocketUpgrader` - Test websocket upgrader setting
  - [ ] `WithEmbedFS` - Test embedded filesystem setting
  - [ ] `WithPublicDir` - Test public directory setting
  - [ ] `WithFormDecoder` - Test form decoder setting
- [ ] Test feature toggles (15 minutes)
  - [ ] `DisableTemplateCache` - Test template cache disabling
  - [ ] `EnableDebugLog` - Test debug logging enabling
  - [ ] `EnableWatch` - Test file watching enabling
  - [ ] `DevelopmentMode` - Test development mode settings
- [ ] **Validation**: Run `./scripts/pre-commit-check.sh --fast` after completion

**Expected Impact**: +15-20% total coverage

### Task 4: Internal Logger Package Tests (20 minutes)

**Priority**: MEDIUM - used throughout codebase  
**Current Coverage**: 0.0% → Target: 80%+

- [ ] Test log level functions
  - [ ] Test `Errorf` function
  - [ ] Test `Infof` function  
  - [ ] Test `Debugf` function
- [ ] Test log formatting
  - [ ] Test message formatting
  - [ ] Test with various argument types
- [ ] **Validation**: Run `./scripts/pre-commit-check.sh --fast` after completion

**Expected Impact**: +3-5% total coverage

### Task 5: Render Function Improvements (60 minutes)

**Priority**: MEDIUM - improve existing coverage  
**Current Coverage**: Improve from 23.8% and 50% → Target: 80%+

- [ ] Improve `uniques` function (30 minutes)
  - [ ] Test with duplicate events (same type, target, key)
  - [ ] Test with nil fields in events
  - [ ] Test edge cases with empty slices
  - [ ] Test complex deduplication scenarios
- [ ] Improve `buildDOMEventFromTemplate` (30 minutes)
  - [ ] Test with special template name "-"
  - [ ] Test with error states and error handling
  - [ ] Test with missing template data
  - [ ] Test template execution errors
- [ ] **Validation**: Run `./scripts/pre-commit-check.sh --fast` after completion

**Expected Impact**: +8-12% total coverage

### Task 6: Websocket Core Functions (90 minutes)

**Priority**: HIGH - core functionality with 0% coverage  
**Current Coverage**: Multiple 0.0% functions → Target: 80%+

- [ ] Test utility functions (30 minutes)
  - [ ] `eqBytesHash` - Test byte hash comparison
  - [ ] `writeConn` - Test writing to websocket connection
  - [ ] `writeEvent` - Test writing events to websocket
- [ ] Test authorization (30 minutes)
  - [ ] `RedirectUnauthorisedWebSocket` - Test redirect behavior
  - [ ] Test proper HTTP response codes
  - [ ] Test redirect headers
- [ ] Improve `onWebsocket` coverage (30 minutes)
  - [ ] Test websocket upgrade scenarios
  - [ ] Test authentication edge cases
  - [ ] Test connection cleanup
- [ ] **Validation**: Run `./scripts/pre-commit-check.sh` (full mode) after completion

**Expected Impact**: +10-15% total coverage

### Task 7: Internal Helper Package Tests (15 minutes)

**Priority**: LOW - test utilities  
**Current Coverage**: 0.0% → Target: 80%+

- [ ] Test helper functions for HTML comparison
  - [ ] Test node comparison functions
  - [ ] Test HTML parsing utilities
- [ ] **Validation**: Run `./scripts/pre-commit-check.sh --fast` after completion

**Expected Impact**: +2-3% total coverage

## Implementation Order (Highest ROI First)

1. **Task 1** (30 min) - Internal session tests → +8-10% coverage
2. **Task 3** (45 min) - Controller options → +15-20% coverage  
3. **Task 6** (90 min) - Websocket functions → +10-15% coverage
4. **Task 5** (60 min) - Render improvements → +8-12% coverage
5. **Task 2** (15 min) - Internal template tests → +3-5% coverage
6. **Task 4** (20 min) - Internal logger tests → +3-5% coverage
7. **Task 7** (15 min) - Internal helper tests → +2-3% coverage

**Total Time**: ~4.5 hours  
**Expected Final Coverage**: 75-85% (target achieved)

## Validation Strategy

- **Fast validation** after each task: `./scripts/pre-commit-check.sh --fast`
- **Full validation** after high-impact tasks: `./scripts/pre-commit-check.sh`
- **Coverage check** after each task: `go test -coverprofile=coverage.out ./...`
- **Continuous flow**: If no changes to non-test files and pre-commit checks pass, continue immediately to the next task without stopping

## Success Criteria

- [ ] **Total Coverage**: 24% → 80%+
- [ ] **All tasks completed** within 6-8 hours
- [ ] **All tests passing** with pre-commit validation
- [ ] **No regressions** in existing functionality
- [ ] **Continuous execution**: Move to next task immediately if validation passes and only test files were modified

## Next Steps

Start with **Task 1 (Internal Session Package)** as it has the highest impact and is used by multiple core components.

**Execution Flow**: After completing each task, run the validation checks. If only test files were modified and pre-commit validation passes, continue immediately to the next task. This allows for rapid iteration through the test coverage improvements without unnecessary stops.

## Testing Strategy Guidelines

### For Each Test

1. **Arrange**: Set up test data and mocks
2. **Act**: Execute the function under test
3. **Assert**: Verify expected outcomes

### Key Testing Patterns

- Use table-driven tests for multiple scenarios
- Mock external dependencies (HTTP, websocket, file system)
- Test both success and error paths
- Test edge cases (nil values, empty data, etc.)
- Use testify/assert for cleaner assertions

### Coverage Validation

- Run `go test -coverprofile=coverage.out ./...` after each task
- Use `go tool cover -func=coverage.out` to verify improvements
- Aim for 80%+ line coverage in each package

## Notes

- Focus on testing public APIs and critical paths first
- Mock external dependencies to ensure isolated testing
- Follow existing test patterns in the codebase
- Ensure tests are maintainable and readable
- Document complex test scenarios for future reference
