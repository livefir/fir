#!/bin/bash

# Fir Framework Pre-Commit Quality Gates
# This script validates that all quality gates pass before committing changes
# Usage: ./scripts/pre-commit-check.sh [--help] [--fast]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LOG_FILE="pre-commit-check-$(date +%Y%m%d-%H%M%S).log"
FAST_MODE=false
USE_LOCAL_CHROME=false

# Functions
show_help() {
    echo "Fir Framework Pre-Commit Quality Gates"
    echo ""
    echo "USAGE:"
    echo "  ./scripts/pre-commit-check.sh [--help] [--fast] [--local-chrome]"
    echo ""
    echo "OPTIONS:"
    echo "  --help          Show this help message"
    echo "  --fast          Enable fast mode - skips coverage analysis and example builds"
    echo "  --local-chrome  Use local Chrome instead of Docker Chrome (requires Chrome/Chromium installed)"
    echo ""
    echo "DESCRIPTION:"
    echo "  This script runs comprehensive validation with aggressive performance optimizations:"
    echo "  - Parallel build compilation with caching and memory optimization"
    echo "  - Ultra-parallel tests (up to 90% CPU utilization) with selective running"
    echo "  - Parallel static analysis (Go Vet + StaticCheck simultaneously)"
    echo "  - Docker Chrome tests by default (fallback to local Chrome with --local-chrome flag)"
    echo "  - Go modules validation"
    echo "  - Alpine.js plugin testing (if changes detected, with smart dependency caching)"
    echo "  - Parallel example compilation (skipped in --fast mode)"
    echo "  - Smart test coverage analysis (skipped in --fast mode)"
    echo "  - Aggressive cleanup of temporary files"
    echo ""
    echo "  PERFORMANCE OPTIMIZATIONS:"
    echo "  - Uses up to 90% of CPU cores for maximum parallel execution"
    echo "  - Intelligent selective testing based on changed files (fast mode)"
    echo "  - Build caching and memory optimization flags"
    echo "  - Parallel static analysis execution"
    echo "  - Reduced timeouts and aggressive garbage collection"
    echo "  - Parallel example compilation with batching"
    echo ""
    echo "  Tests run in parallel using up to 90% of available CPU cores for maximum speed."
    echo "  Chrome tests use Docker by default for consistency and isolation."
    echo "  Use --local-chrome flag if you prefer to use locally installed Chrome/Chromium."
    echo "  Returns exit code 0 if all quality gates pass, non-zero otherwise."
    echo ""
    echo "EXAMPLES:"
    echo "  ./scripts/pre-commit-check.sh                    # Full validation with Docker Chrome"
    echo "  ./scripts/pre-commit-check.sh --fast             # Fast validation with Docker Chrome"
    echo "  ./scripts/pre-commit-check.sh --local-chrome     # Full validation with local Chrome"
    echo "  ./scripts/pre-commit-check.sh --help             # Show this help"
    echo ""
    echo "  # Use in other scripts:"
    echo "  if ./scripts/pre-commit-check.sh; then"
    echo "    echo 'All quality gates passed!'"
    echo "  else"
    echo "    echo 'Quality gates failed'"
    echo "  fi"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --fast|-f)
            FAST_MODE=true
            shift
            ;;
        --local-chrome)
            USE_LOCAL_CHROME=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}✅ $1${NC}" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}❌ $1${NC}" | tee -a "$LOG_FILE"
}

header() {
    echo -e "\n${BLUE}===========================================${NC}" | tee -a "$LOG_FILE"
    echo -e "${BLUE}$1${NC}" | tee -a "$LOG_FILE" 
    echo -e "${BLUE}===========================================${NC}" | tee -a "$LOG_FILE"
}

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local success_msg="$3"
    local error_msg="$4"
    
    log "Running: $test_name"
    log "Command: $test_cmd"
    
    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        success "$success_msg"
        return 0
    else
        error "$error_msg"
        echo "Check $LOG_FILE for detailed error output"
        return 1
    fi
}

cleanup_temp_files() {
    log "Cleaning up temporary files..."
    
    # Remove common temporary test files
    find . -name "*.test" -type f -delete 2>/dev/null || true
    find . -name "*.out" -type f -delete 2>/dev/null || true
    find . -name "coverage.out" -type f -delete 2>/dev/null || true
    find . -name "profile.out" -type f -delete 2>/dev/null || true
    find . -name "cpu.prof" -type f -delete 2>/dev/null || true
    find . -name "mem.prof" -type f -delete 2>/dev/null || true
    
    # Only remove old pre-commit log files that are more than 24 hours old
    # This preserves recent logs for debugging while cleaning up very old ones
    # Keep current log file and any recent ones for debugging failures
    find . -name "pre-commit-check-*.log" ! -name "$LOG_FILE" -type f -mtime +1 -delete 2>/dev/null || true
    
    # Remove database files that might be created during tests
    find . -name "*.db" -type f -delete 2>/dev/null || true
    find . -name "test.sqlite" -type f -delete 2>/dev/null || true
    
    # Remove any build artifacts in examples
    if [ -d "examples" ]; then
        find examples -name "main" -type f -delete 2>/dev/null || true
        find examples -name "*.exe" -type f -delete 2>/dev/null || true
    fi
    
    # Remove any backup files
    find . -name "*.backup" -type f -delete 2>/dev/null || true
    
    # Remove Node.js/Alpine.js plugin artifacts (but preserve node_modules)
    if [ -d "alpinejs-plugin" ]; then
        # Only remove test/build artifacts, not dependencies
        rm -rf alpinejs-plugin/coverage 2>/dev/null || true
        rm -rf alpinejs-plugin/.nyc_output 2>/dev/null || true
        # Note: Preserving node_modules and package-lock.json for faster builds
    fi
    
    success "Temporary files cleaned up"
}

# Run Chrome-based tests (sanity, e2e) with Docker support
run_chrome_tests() {
    # Check if there are Chrome-based tests to run
    if [ ! -d "examples/sanity" ] && [ ! -d "examples/e2e" ]; then
        return 0
    fi
    
    # Skip Chrome tests in fast mode
    if [ "$FAST_MODE" = true ]; then
        log "Skipping Chrome-based tests in fast mode"
        return 0  # Don't require Chrome tests in fast mode
    fi
    
    header "🌐 Chrome-Based Integration Tests"
    
    # Docker Chrome approach (default)
    if [ "$USE_LOCAL_CHROME" != true ]; then
        if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
            log "Using Docker Chrome (default approach)"
            
            # Check if chrome container is already running
            if ! docker ps | grep -q "chrome-test"; then
                log "Starting Chrome Docker container..."
                if docker run --rm -d --name chrome-test -p 9222:9222 chromedp/headless-shell:latest >/dev/null 2>&1; then
                    # Wait for Chrome to be ready
                    log "Waiting for Chrome container to be ready..."
                    sleep 3
                    
                    # Verify Chrome is accessible
                    if curl -s http://localhost:9222/json/version >/dev/null 2>&1; then
                        log "Chrome container ready - running tests"
                        if CHROME_REMOTE_URL=ws://localhost:9222 run_test "Chrome Tests (Docker)" \
                                                                          "go test -timeout=2m ./examples/sanity/ ./examples/e2e/" \
                                                                          "Chrome tests passed with Docker" \
                                                                          "Chrome tests failed with Docker"; then
                            docker stop chrome-test >/dev/null 2>&1 || true
                            return 0
                        fi
                        docker stop chrome-test >/dev/null 2>&1 || true
                    else
                        warning "Chrome container failed to start properly"
                        docker stop chrome-test >/dev/null 2>&1 || true
                    fi
                else
                    warning "Failed to start Chrome Docker container"
                fi
            else
                log "Chrome container already running - running tests"
                if CHROME_REMOTE_URL=ws://localhost:9222 run_test "Chrome Tests (Docker)" \
                                                                  "go test -timeout=2m ./examples/sanity/ ./examples/e2e/" \
                                                                  "Chrome tests passed with Docker" \
                                                                  "Chrome tests failed with Docker"; then
                    return 0
                fi
            fi
        else
            warning "Docker not available for Chrome testing"
        fi
    fi
    
    # Local Chrome approach (when explicitly requested or Docker failed)
    if [ "$USE_LOCAL_CHROME" = true ]; then
        CHROME_EXECUTABLE=""
        if command -v google-chrome >/dev/null 2>&1; then
            CHROME_EXECUTABLE="google-chrome"
        elif command -v chromium >/dev/null 2>&1; then
            CHROME_EXECUTABLE="chromium"
        elif [ -f "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" ]; then
            CHROME_EXECUTABLE="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
        fi
        
        if [ -n "$CHROME_EXECUTABLE" ]; then
            log "Using local Chrome as requested: $CHROME_EXECUTABLE"
            if run_test "Chrome Tests (Local)" \
                        "go test -timeout=2m ./examples/sanity/ ./examples/e2e/" \
                        "Chrome tests passed with local Chrome" \
                        "Chrome tests failed with local Chrome"; then
                return 0
            fi
        fi
    fi
    
    # If we get here, all Chrome test approaches failed
    error "Chrome-based tests FAILED: Docker Chrome or local Chrome is required for quality gates"
    error ""
    error "RECOMMENDED SOLUTION: Install Docker for consistent testing:"
    error "  # macOS"
    error "  brew install --cask docker"
    error "  # OR for local Chrome approach, use --local-chrome flag with:"
    error "  brew install --cask google-chrome"
    error "  # OR"
    error "  brew install --cask chromium"
    error ""
    error "Chrome-based tests are mandatory integration tests that validate:"
    error "  - Browser automation functionality"
    error "  - Race condition handling"
    error "  - End-to-end user workflows"
    error ""
    error "Docker Chrome is preferred for consistency and isolation."
    error "Use --local-chrome flag if you prefer using locally installed Chrome."
    
    return 1  # Fail the entire pre-commit - Chrome tests are mandatory
}

# Main execution
main() {
    # Record start time for performance tracking
    START_TIME=$(date +%s)
    
    if [ "$FAST_MODE" = true ]; then
        header "🚀 Fir Framework Pre-Commit Quality Gates (FAST MODE)"
        log "Fast mode enabled - skipping coverage analysis and example builds"
    else
        header "🚀 Fir Framework Pre-Commit Quality Gates"
    fi
    
    if [ "$USE_LOCAL_CHROME" = true ]; then
        log "Local Chrome mode enabled - will use locally installed Chrome instead of Docker"
    else
        log "Using Docker Chrome (default) - use --local-chrome flag for local Chrome"
    fi
    
    log "Starting comprehensive validation..."
    log "Log file: $LOG_FILE"
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
        exit 1
    fi
    
    # Clean up any temporary test files first
    header "🧹 Cleanup Temporary Files"
    cleanup_temp_files
    
    FAILED_TESTS=0
    
    # 1. Build Validation (with parallel compilation, caching, and memory optimization)
    header "🔨 Build Validation"
    
    # Use parallel compilation for faster builds with optimizations
    CPU_COUNT_BUILD=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
    BUILD_PARALLELISM=$((CPU_COUNT_BUILD))
    
    # Enable build cache and memory optimizations
    export GOCACHE=$(go env GOCACHE 2>/dev/null || echo "$HOME/.cache/go-build")
    
    if [ "$FAST_MODE" = true ]; then
        # Fast mode: aggressive caching, no race detection, optimized flags
        BUILD_FLAGS="-p $BUILD_PARALLELISM -ldflags='-s -w' -trimpath"
        BUILD_LABEL="Go Build (Fast & Cached)"
        log "Fast mode: Using aggressive build optimizations with caching"
    else
        # Normal mode: balanced approach with caching
        BUILD_FLAGS="-p $BUILD_PARALLELISM -trimpath"
        BUILD_LABEL="Go Build (Parallel & Cached)"
        log "Using parallel build with caching and optimization"
    fi
    
    if ! run_test "$BUILD_LABEL" \
                  "go build $BUILD_FLAGS ./..." \
                  "Optimized build completed successfully (using $BUILD_PARALLELISM processes)" \
                  "Optimized build failed - check for compilation errors"; then
        ((FAILED_TESTS++))
    fi
    
    # 2. Tests (with enhanced parallelism and selective running)
    header "🧪 Tests"
    
    # Determine number of parallel processes - use up to 90% of CPU cores for speed
    CPU_COUNT=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
    
    # Check if we can do selective testing based on changed files
    CHANGED_FILES=$(git diff --name-only HEAD~1..HEAD 2>/dev/null || echo "")
    STAGED_FILES=$(git diff --staged --name-only 2>/dev/null || echo "")
    ALL_CHANGED="$CHANGED_FILES $STAGED_FILES"
    
    # Determine if we need full test suite or can run selective tests
    NEEDS_FULL_TESTS=true
    if [ -n "$ALL_CHANGED" ] && [ "$FAST_MODE" = true ]; then
        # Check if changes are only in specific packages
        GO_CHANGES=$(echo "$ALL_CHANGED" | grep -E '\.(go|mod|sum)$' || true)
        if [ -n "$GO_CHANGES" ]; then
            # Extract package directories from changed files
            CHANGED_PACKAGES=$(echo "$GO_CHANGES" | grep '\.go$' | xargs -I {} dirname {} | sort -u | grep -v '^examples/e2e' || true)
            if [ -n "$CHANGED_PACKAGES" ] && [ $(echo "$CHANGED_PACKAGES" | wc -l) -le 3 ]; then
                NEEDS_FULL_TESTS=false
                log "Detected changes in limited packages: $(echo $CHANGED_PACKAGES | tr '\n' ' ')"
            fi
        fi
    fi
    
    if [ "$FAST_MODE" = true ]; then
        # In fast mode, use very aggressive parallelism for maximum speed
        PARALLEL_JOBS=$((CPU_COUNT * 9 / 10))
        if [ $PARALLEL_JOBS -lt 4 ]; then
            PARALLEL_JOBS=4
        fi
        # Cap at higher maximum for fast mode
        if [ $PARALLEL_JOBS -gt 20 ]; then
            PARALLEL_JOBS=20
        fi
        TEST_FLAGS="-parallel $PARALLEL_JOBS -short -timeout=20s -count=1"
        log "Fast mode: Using $PARALLEL_JOBS parallel processes with 20s timeout (out of $CPU_COUNT CPUs)"
        COVERAGE_ANALYSIS=false
    else
        # In normal mode, use aggressive parallelism but leave room for coverage processing
        PARALLEL_JOBS=$((CPU_COUNT * 4 / 5))
        if [ $PARALLEL_JOBS -lt 3 ]; then
            PARALLEL_JOBS=3
        fi
        # Cap at reasonable maximum for coverage mode - reduced to avoid race conditions
        if [ $PARALLEL_JOBS -gt 3 ]; then
            PARALLEL_JOBS=3
        fi
        TEST_FLAGS="-parallel $PARALLEL_JOBS -count=1 -coverprofile=coverage.out -timeout=2m"
        log "Using $PARALLEL_JOBS parallel test processes with coverage and 2m timeout (out of $CPU_COUNT CPUs)"
        COVERAGE_ANALYSIS=true
    fi
    
    # Set memory optimizations for tests
    export GOMAXPROCS=$PARALLEL_JOBS
    export GOGC=100  # More aggressive garbage collection
    
    # Optimize test command based on environment and changes
    if [ "$FAST_MODE" = true ]; then
        if [ "$NEEDS_FULL_TESTS" = false ] && [ -n "$CHANGED_PACKAGES" ]; then
            # Selective testing for fast mode
            TEST_LABEL="Tests (Selective - Fast Mode)"
            PACKAGE_LIST=$(echo "$CHANGED_PACKAGES" | sed 's|^|./|' | tr '\n' ' ')
            TEST_CMD="go test $TEST_FLAGS $PACKAGE_LIST"
            log "Fast mode: Running selective tests for changed packages only"
        else
            # Full fast mode excluding slow tests
            TEST_LABEL="Tests (Fast Mode - Ultra High Parallelism)"
            TEST_CMD="go test $TEST_FLAGS \$(go list ./... | grep -v -E \"examples/e2e|integration_test\")"
            log "Fast mode: Excluding slow e2e and integration tests, using $PARALLEL_JOBS parallel processes"
        fi
    else
        # Check if Docker is running and use appropriate test command
        if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
            TEST_LABEL="Tests with Docker (Excluding E2E and Sanity)"
            TEST_CMD="DOCKER=1 go test $TEST_FLAGS \$(go list ./... | grep -v -E \"examples/e2e|examples/sanity\")"
            log "Docker available: Running full test suite excluding E2E and sanity tests with $PARALLEL_JOBS parallel processes"
        else
            TEST_LABEL="Tests (Excluding E2E and Sanity)"
            TEST_CMD="go test $TEST_FLAGS \$(go list ./... | grep -v -E \"examples/e2e|examples/sanity\")"
            log "Running full test suite excluding E2E and sanity tests with $PARALLEL_JOBS parallel processes"
        fi
    fi
    
    if ! run_test "$TEST_LABEL" \
                  "$TEST_CMD" \
                  "All tests passed" \
                  "Tests failed"; then
        ((FAILED_TESTS++))
    else
        # Process coverage results if enabled
        if [ "$COVERAGE_ANALYSIS" = true ] && [ -f coverage.out ]; then
            log "Processing coverage results..."
            COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')
            if [ -n "$COVERAGE" ] && (( $(echo "$COVERAGE < 50" | bc -l) )); then
                warning "Test coverage is below 50% ($COVERAGE%)"
            else
                success "Test coverage: $COVERAGE%"
            fi
            rm -f coverage.out
        fi
    fi
    
    # Run Chrome-based tests if conditions are met
    if ! run_chrome_tests; then
        ((FAILED_TESTS++))
    fi
    
    # 3. Parallel Static Analysis (Go Vet + StaticCheck)
    header "🔍 Static Analysis (Parallel)"
    
    # Run static analysis tools in parallel for speed
    VET_RESULT=0
    STATICCHECK_RESULT=0
    
    # Create temporary files for parallel execution
    VET_LOG=$(mktemp)
    STATICCHECK_LOG=$(mktemp)
    
    log "Running Go Vet and StaticCheck in parallel for faster analysis"
    
    # Run go vet in background
    (
        if go vet ./... > "$VET_LOG" 2>&1; then
            echo "SUCCESS" >> "$VET_LOG"
        else
            echo "FAILED" >> "$VET_LOG"
        fi
    ) &
    VET_PID=$!
    
    # Run staticcheck in background (if available)
    if command -v staticcheck >/dev/null 2>&1; then
        (
            if staticcheck ./... > "$STATICCHECK_LOG" 2>&1; then
                echo "SUCCESS" >> "$STATICCHECK_LOG"
            else
                echo "FAILED" >> "$STATICCHECK_LOG"
            fi
        ) &
        STATICCHECK_PID=$!
        STATICCHECK_AVAILABLE=true
    else
        STATICCHECK_AVAILABLE=false
        warning "StaticCheck not installed - install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
    fi
    
    # Wait for go vet to complete
    wait $VET_PID
    if grep -q "SUCCESS" "$VET_LOG"; then
        success "Go vet analysis passed"
        # Add output to log, excluding SUCCESS marker (use sed instead of head -n -1 for portability)
        sed '$d' "$VET_LOG" >> "$LOG_FILE"
    else
        error "Go vet found issues"
        # Add output to log, excluding FAILED marker (use sed instead of head -n -1 for portability)
        sed '$d' "$VET_LOG" >> "$LOG_FILE"
        ((FAILED_TESTS++))
    fi
    rm "$VET_LOG"
    
    # Wait for staticcheck to complete (if running)
    if [ "$STATICCHECK_AVAILABLE" = true ]; then
        wait $STATICCHECK_PID
        if grep -q "SUCCESS" "$STATICCHECK_LOG"; then
            success "StaticCheck analysis passed"
            # Add output to log, excluding SUCCESS marker (use sed instead of head -n -1 for portability)
            sed '$d' "$STATICCHECK_LOG" >> "$LOG_FILE"
        else
            error "StaticCheck found issues"
            # Add output to log, excluding FAILED marker (use sed instead of head -n -1 for portability)
            sed '$d' "$STATICCHECK_LOG" >> "$LOG_FILE"
            ((FAILED_TESTS++))
        fi
        rm "$STATICCHECK_LOG"
    fi
    
    # 5. Go Mod Tidy Check
    header "📦 Go Modules Validation"
    cp go.mod go.mod.backup
    cp go.sum go.sum.backup
    
    if ! run_test "Go Mod Tidy" \
                  "go mod tidy" \
                  "Go modules are tidy" \
                  "Go mod tidy made changes"; then
        # Restore original files if go mod tidy changed them
        mv go.mod.backup go.mod
        mv go.sum.backup go.sum
        ((FAILED_TESTS++))
    else
        # Check if files changed
        if ! diff go.mod go.mod.backup >/dev/null || ! diff go.sum go.sum.backup >/dev/null; then
            warning "go mod tidy made changes - please commit them"
            ((FAILED_TESTS++))
        fi
        rm go.mod.backup go.sum.backup
    fi
    
    # 6. Alpine.js Plugin Testing (if plugin directory exists and has changes)
    header "🌲 Alpine.js Plugin Testing"
    if [ -d "alpinejs-plugin" ]; then
        # Check if there are any changes in the plugin directory (committed or staged)
        PLUGIN_CHANGES=$(git diff HEAD~1..HEAD --name-only | grep "^alpinejs-plugin/" || true)
        STAGED_PLUGIN_CHANGES=$(git diff --staged --name-only | grep "^alpinejs-plugin/" || true)
        
        if [ -n "$PLUGIN_CHANGES" ] || [ -n "$STAGED_PLUGIN_CHANGES" ]; then
            log "Alpine.js plugin changes detected - running plugin tests..."
            
            # Check if Node.js and npm are available
            if ! command -v node >/dev/null 2>&1 || ! command -v npm >/dev/null 2>&1; then
                error "Node.js and npm are required for Alpine.js plugin testing"
                ((FAILED_TESTS++))
            else
                cd alpinejs-plugin
                
                # Check if dependencies need to be installed or updated
                NEED_INSTALL=false
                if [ ! -d "node_modules" ] || [ ! -f "package-lock.json" ]; then
                    log "Node modules not found - installing dependencies..."
                    NEED_INSTALL=true
                elif [ "package.json" -nt "node_modules" ]; then
                    log "package.json is newer than node_modules - updating dependencies..."
                    NEED_INSTALL=true
                else
                    log "Dependencies are up to date - skipping npm install"
                fi
                
                # Install dependencies only if needed
                if [ "$NEED_INSTALL" = true ]; then
                    if ! run_test "Alpine.js Plugin Dependencies" \
                                  "npm install" \
                                  "Dependencies installed successfully" \
                                  "Failed to install dependencies"; then
                        ((FAILED_TESTS++))
                        cd ..
                    else
                        # Run plugin tests
                        if ! run_test "Alpine.js Plugin Tests" \
                                      "npm test" \
                                      "All Alpine.js plugin tests passed" \
                                      "Alpine.js plugin tests failed"; then
                            ((FAILED_TESTS++))
                        fi
                        
                        # Run plugin build
                        if ! run_test "Alpine.js Plugin Build" \
                                      "npm run build" \
                                      "Plugin build completed successfully" \
                                      "Plugin build failed"; then
                            ((FAILED_TESTS++))
                        fi
                        
                        cd ..
                    fi
                else
                    # Dependencies are up to date, run tests directly
                    if ! run_test "Alpine.js Plugin Tests" \
                                  "npm test" \
                                  "All Alpine.js plugin tests passed" \
                                  "Alpine.js plugin tests failed"; then
                        ((FAILED_TESTS++))
                    fi
                    
                    # Run plugin build
                    if ! run_test "Alpine.js Plugin Build" \
                                  "npm run build" \
                                  "Plugin build completed successfully" \
                                  "Plugin build failed"; then
                        ((FAILED_TESTS++))
                    fi
                    
                    cd ..
                fi
            fi
        else
            log "No changes detected in alpinejs-plugin directory - skipping plugin tests"
        fi
    else
        log "Alpine.js plugin directory not found - skipping plugin tests"
    fi
    
    # 7. Parallel Example Compilation Check (skip in fast mode)
    if [ "$FAST_MODE" != true ]; then
        header "📚 Example Compilation Check (Parallel)"
        if [ -d "examples" ]; then
            log "Compiling examples in parallel for faster validation"
            
            # Create array of example directories (excluding unwanted examples)
            EXAMPLE_DIRS=()
            for example_dir in examples/*/; do
                example_name=$(basename "$example_dir")
                # Skip unwanted examples
                if [ "$example_name" = "custom_template_engine" ] || [ "$example_name" = "template_engine_example" ]; then
                    log "Skipping excluded example: $example_name"
                    continue
                fi
                if [ -f "$example_dir/main.go" ]; then
                    EXAMPLE_DIRS+=("$example_dir")
                fi
            done
            
            if [ ${#EXAMPLE_DIRS[@]} -gt 0 ]; then
                # Determine parallelism for examples (use fewer processes than tests)
                EXAMPLE_PARALLELISM=$((CPU_COUNT / 2))
                if [ $EXAMPLE_PARALLELISM -lt 1 ]; then
                    EXAMPLE_PARALLELISM=1
                fi
                if [ $EXAMPLE_PARALLELISM -gt 6 ]; then
                    EXAMPLE_PARALLELISM=6  # Cap to avoid too many processes
                fi
                
                log "Building ${#EXAMPLE_DIRS[@]} examples using $EXAMPLE_PARALLELISM parallel processes"
                
                # Build examples in parallel batches
                EXAMPLE_FAILED=0
                PIDS=()
                TEMP_RESULTS=()
                
                for ((i=0; i<${#EXAMPLE_DIRS[@]}; i++)); do
                    example_dir="${EXAMPLE_DIRS[$i]}"
                    example_name=$(basename "$example_dir")
                    result_file=$(mktemp)
                    TEMP_RESULTS+=("$result_file")
                    
                    # Start build in background
                    (
                        if (cd "$example_dir" && go build . > "$result_file.out" 2>&1); then
                            echo "SUCCESS:$example_name" > "$result_file"
                        else
                            echo "FAILED:$example_name" > "$result_file"
                            cat "$result_file.out" >> "$result_file"
                        fi
                        rm -f "$result_file.out"
                    ) &
                    
                    PIDS+=($!)
                    
                    # If we've started max parallel processes, wait for one to complete
                    if [ ${#PIDS[@]} -ge $EXAMPLE_PARALLELISM ]; then
                        wait ${PIDS[0]}
                        PIDS=("${PIDS[@]:1}")  # Remove first PID
                    fi
                done
                
                # Wait for all remaining processes
                for pid in "${PIDS[@]}"; do
                    wait $pid
                done
                
                # Collect results
                for result_file in "${TEMP_RESULTS[@]}"; do
                    if [ -f "$result_file" ]; then
                        result_line=$(head -n 1 "$result_file")
                        example_name=$(echo "$result_line" | cut -d: -f2)
                        
                        if echo "$result_line" | grep -q "SUCCESS"; then
                            success "Example $example_name builds successfully"
                        else
                            error "Example $example_name failed to build"
                            # Add error details to log
                            tail -n +2 "$result_file" >> "$LOG_FILE"
                            ((EXAMPLE_FAILED++))
                        fi
                        rm -f "$result_file"
                    fi
                done
                
                if [ $EXAMPLE_FAILED -eq 0 ]; then
                    success "All ${#EXAMPLE_DIRS[@]} examples compiled successfully in parallel"
                else
                    error "$EXAMPLE_FAILED out of ${#EXAMPLE_DIRS[@]} examples failed to compile"
                    ((FAILED_TESTS++))
                fi
            else
                log "No example directories with main.go found"
            fi
        fi
    else
        log "Skipping example compilation check in fast mode"
    fi
    
    # Final Results
    header "📋 Quality Gate Results"
    
    # Calculate execution time
    END_TIME=$(date +%s)
    EXECUTION_TIME=$((END_TIME - START_TIME))
    MINUTES=$((EXECUTION_TIME / 60))
    SECONDS=$((EXECUTION_TIME % 60))
    
    if [ $MINUTES -gt 0 ]; then
        TIME_DISPLAY="${MINUTES}m ${SECONDS}s"
    else
        TIME_DISPLAY="${SECONDS}s"
    fi
    
    # Final cleanup of any temp files created during testing
    cleanup_temp_files
    
    if [ $FAILED_TESTS -eq 0 ]; then
        success "🎉 ALL QUALITY GATES PASSED! (${TIME_DISPLAY})"
        success "✅ Code is ready for commit"
        
        # Performance feedback
        if [ "$FAST_MODE" = true ]; then
            log "Fast mode completed in ${TIME_DISPLAY}"
            log "Note: e2e tests were skipped for speed. Run full validation before committing critical changes."
        else
            log "Full validation completed in ${TIME_DISPLAY}"
            if [ $EXECUTION_TIME -gt 30 ]; then
                log "💡 Tip: Use --fast flag for quicker validation during development"
            fi
        fi
        
        # Only clean up log file if all tests passed
        rm "$LOG_FILE" 2>/dev/null || true
        success "Log file cleaned up (all tests passed)"
        
        # Mark script as successful to prevent cleanup on exit
        SCRIPT_SUCCESS="true"
        
        exit 0
    else
        error "❌ $FAILED_TESTS quality gate(s) failed (${TIME_DISPLAY})"
        error "Code is NOT ready for commit"
        echo ""
        echo "Please fix the failing tests and run the validation script again."
        echo "Detailed logs are available in: $LOG_FILE"
        echo "💡 Log file preserved for debugging: $LOG_FILE"
        exit 1
    fi
}

# Trap to clean up on exit
cleanup() {
    if [ -f "go.mod.backup" ]; then
        mv go.mod.backup go.mod
    fi
    if [ -f "go.sum.backup" ]; then
        mv go.sum.backup go.sum
    fi
    
    # Clean up any remaining temporary files
    cleanup_temp_files 2>/dev/null || true
    
    # Only clean up the current log file if the script completed successfully
    # If there were failures or unexpected exits, preserve the log for debugging
    if [ -f "$LOG_FILE" ] && [ "${SCRIPT_SUCCESS:-}" != "true" ]; then
        echo "Log file preserved for debugging: $LOG_FILE"
    fi
}
trap cleanup EXIT

# Run main function
main "$@"
