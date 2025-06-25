#!/bin/bash

# Fir Framework Pre-Commit Quality Gates
# This script validates that all quality gates pass before committing changes
# Usage: ./scripts/pre-commit-check.sh [--help]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LOG_FILE="pre-commit-check-$(date +%Y%m%d-%H%M%S).log"

# Functions
show_help() {
    echo "Fir Framework Pre-Commit Quality Gates"
    echo ""
    echo "USAGE:"
    echo "  ./scripts/pre-commit-check.sh [--help]"
    echo ""
    echo "OPTIONS:"
    echo "  --help      Show this help message"
    echo ""
    echo "DESCRIPTION:"
    echo "  This script runs comprehensive validation including:"
    echo "  - Build compilation (go build ./...)"
    echo "  - Docker environment tests (DOCKER=1 go test ./...) - if Docker is available"
    echo "  - Basic tests (go test ./...) - if Docker is not available"
    echo "  - Static analysis (go vet, staticcheck)"
    echo "  - Go modules validation"
    echo "  - Alpine.js plugin testing (if changes detected)"
    echo "  - Example compilation check"
    echo "  - Cleanup of temporary files"
    echo ""
    echo "  Docker tests are automatically skipped if Docker is not installed or running."
    echo "  Returns exit code 0 if all quality gates pass, non-zero otherwise."
    echo ""
    echo "EXAMPLES:"
    echo "  ./scripts/pre-commit-check.sh"
    echo "  ./scripts/pre-commit-check.sh --help"
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
    find . -name "*.log" -path "./pre-commit-check-*.log" -prune -o -name "*.log" -type f -delete 2>/dev/null || true
    
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
    
    # Remove Node.js/Alpine.js plugin artifacts
    if [ -d "alpinejs-plugin" ]; then
        rm -rf alpinejs-plugin/node_modules 2>/dev/null || true
        rm -f alpinejs-plugin/package-lock.json 2>/dev/null || true
        rm -rf alpinejs-plugin/coverage 2>/dev/null || true
        rm -rf alpinejs-plugin/.nyc_output 2>/dev/null || true
    fi
    
    success "Temporary files cleaned up"
}

# Main execution
main() {
    header "🚀 Fir Framework Pre-Commit Quality Gates"
    
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
    
    # 1. Build Validation
    header "🔨 Build Validation"
    if ! run_test "Go Build" \
                  "go build ./..." \
                  "Build completed successfully" \
                  "Build failed - check for compilation errors"; then
        ((FAILED_TESTS++))
    fi
    
    # 2. Docker Environment Tests (includes core tests)
    header "🐳 Docker Environment Tests"
    
    # Check if Docker is available and running
    if ! command -v docker >/dev/null 2>&1; then
        warning "Docker is not installed - skipping Docker tests"
        log "Install Docker to run full test suite: https://docs.docker.com/get-docker/"
        
        # Run basic tests without Docker (excluding e2e tests that may hang)
        header "🧪 Basic Tests (without Docker)"
        if ! run_test "Basic Tests" \
                      "timeout 300 bash -c 'go test \$(go list ./... | grep -v /examples/e2e)'" \
                      "All basic tests passed" \
                      "Basic tests failed"; then
            ((FAILED_TESTS++))
        fi
    elif ! docker info >/dev/null 2>&1; then
        warning "Docker is not running - skipping Docker tests"
        log "Start Docker daemon to run full test suite: sudo systemctl start docker (Linux) or start Docker Desktop (Mac/Windows)"
        
        # Run basic tests without Docker (excluding e2e tests that may hang)
        header "🧪 Basic Tests (without Docker)"
        if ! run_test "Basic Tests" \
                      "timeout 300 bash -c 'go test \$(go list ./... | grep -v /examples/e2e)'" \
                      "All basic tests passed" \
                      "Basic tests failed"; then
            ((FAILED_TESTS++))
        fi
    else
        if ! run_test "Docker Tests" \
                      "timeout 600 bash -c 'DOCKER=1 go test ./...'" \
                      "All Docker environment tests passed" \
                      "Docker environment tests failed or timed out"; then
            ((FAILED_TESTS++))
        fi
    fi
    
    # 3. Static Analysis - Go Vet
    header "🔍 Static Analysis - Go Vet"
    if ! run_test "Go Vet" \
                  "go vet ./..." \
                  "Go vet analysis passed" \
                  "Go vet found issues"; then
        ((FAILED_TESTS++))
    fi
    
    # 4. Static Analysis - StaticCheck (if available)
    header "🔍 Static Analysis - StaticCheck"
    if command -v staticcheck >/dev/null 2>&1; then
        if ! run_test "StaticCheck" \
                      "staticcheck ./..." \
                      "StaticCheck analysis passed" \
                      "StaticCheck found issues"; then
            ((FAILED_TESTS++))
        fi
    else
        warning "StaticCheck not installed - install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
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
    
    # 6. Test Coverage Check (Optional)
    header "📊 Test Coverage Analysis"
    if run_test "Coverage Analysis" \
               "go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out" \
               "Coverage analysis completed" \
               "Coverage analysis failed"; then
        # Extract overall coverage percentage
        COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$COVERAGE < 50" | bc -l) )); then
            warning "Test coverage is below 50% ($COVERAGE%)"
        else
            success "Test coverage: $COVERAGE%"
        fi
        rm -f coverage.out
    fi
    
    # 7. Alpine.js Plugin Testing (if plugin directory exists and has changes)
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
                
                # Install dependencies
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
            fi
        else
            log "No changes detected in alpinejs-plugin directory - skipping plugin tests"
        fi
    else
        log "Alpine.js plugin directory not found - skipping plugin tests"
    fi
    
    # 8. Example Compilation Check
    header "📚 Example Compilation Check"
    if [ -d "examples" ]; then
        EXAMPLE_FAILED=0
        for example_dir in examples/*/; do
            if [ -f "$example_dir/main.go" ]; then
                example_name=$(basename "$example_dir")
                if ! run_test "Example: $example_name" \
                              "cd $example_dir && go build ." \
                              "Example $example_name builds successfully" \
                              "Example $example_name failed to build"; then
                    ((EXAMPLE_FAILED++))
                fi
            fi
        done
        
        if [ $EXAMPLE_FAILED -eq 0 ]; then
            success "All examples compile successfully"
        else
            error "$EXAMPLE_FAILED examples failed to compile"
            ((FAILED_TESTS++))
        fi
    fi
    
    # Final Results
    header "📋 Quality Gate Results"
    
    # Final cleanup of any temp files created during testing
    cleanup_temp_files
    
    if [ $FAILED_TESTS -eq 0 ]; then
        success "🎉 ALL QUALITY GATES PASSED!"
        success "✅ Code is ready for commit"
        
        # Clean up log file if successful
        echo ""
        read -p "Remove log file $LOG_FILE? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            rm "$LOG_FILE"
            success "Log file cleaned up"
        fi
        
        exit 0
    else
        error "❌ $FAILED_TESTS quality gate(s) failed"
        error "Code is NOT ready for commit"
        echo ""
        echo "Please fix the failing tests and run the validation script again."
        echo "Detailed logs are available in: $LOG_FILE"
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
}
trap cleanup EXIT

# Run main function
main "$@"
