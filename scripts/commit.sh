#!/bin/bash

# Fir Framework Commit Script
# This script runs pre-commit checks and creates validated commits
# Usage: ./scripts/commit.sh [commit-message] [--amend]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMMIT_MESSAGE=""
AMEND_MODE=false
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Functions
show_help() {
    echo "Fir Framework Commit Script"
    echo ""
    echo "USAGE:"
    echo "  ./scripts/commit.sh [commit-message] [--amend] [--help]"
    echo ""
    echo "OPTIONS:"
    echo "  --amend     Amend the previous commit instead of creating a new one"
    echo "  --help      Show this help message"
    echo ""
    echo "EXAMPLES:"
    echo "  ./scripts/commit.sh \"Add new feature\""
    echo "  ./scripts/commit.sh \"Fix bug in handler\" --amend"
    echo "  ./scripts/commit.sh --amend"
    echo ""
    echo "DESCRIPTION:"
    echo "  This script creates validated commits by ensuring git pre-commit hooks are installed."
    echo "  The actual validation is handled by the git pre-commit hook which runs:"
    echo "  - Build compilation (go build ./...)"
    echo "  - Docker environment tests (DOCKER=1 go test ./...)"
    echo "  - Static analysis (go vet, staticcheck)"
    echo "  - Go modules validation"
    echo "  - Alpine.js plugin testing (if changes detected)"
    echo "  - Example compilation check"
    echo "  - Cleanup of temporary files"
    echo ""
    echo "  The git hook must be installed for commits to be validated."
}

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --amend)
            AMEND_MODE=true
            shift
            ;;
        *)
            if [ -z "$COMMIT_MESSAGE" ]; then
                COMMIT_MESSAGE="$1"
            else
                error "Multiple commit messages provided. Use quotes for messages with spaces."
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate arguments
if [ "$AMEND_MODE" = false ] && [ -z "$COMMIT_MESSAGE" ]; then
    error "Commit message is required when not amending"
    echo ""
    show_help
    exit 1
fi

# Main execution
main() {
    log "🚀 Starting Fir Framework commit process..."
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
        exit 1
    fi
    
    # Check if there are changes to commit (unless amending)
    if [ "$AMEND_MODE" = false ]; then
        if git diff --staged --quiet && git diff --quiet; then
            warning "No changes detected to commit"
            echo "Use 'git add' to stage changes or '--amend' to amend the previous commit"
            exit 1
        fi
        
        if git diff --staged --quiet; then
            warning "No staged changes detected"
            echo "Use 'git add' to stage changes for commit"
            exit 1
        fi
    fi
    
    # Check if pre-commit hook is installed
    if [ ! -f ".git/hooks/pre-commit" ]; then
        error "Git pre-commit hook is not installed"
        echo "Please run './scripts/install-git-hook.sh' to install the pre-commit hook"
        exit 1
    fi
    
    log "Git pre-commit hook is installed ✓"
    
    # Create the commit (validation will be handled by the git hook)
    if [ "$AMEND_MODE" = true ]; then
        log "Amending previous commit..."
        if [ -n "$COMMIT_MESSAGE" ]; then
            git commit --amend -m "$COMMIT_MESSAGE"
        else
            git commit --amend --no-edit
        fi
        success "Commit amended successfully!"
    else
        log "Creating new commit with message: '$COMMIT_MESSAGE'"
        git commit -m "$COMMIT_MESSAGE"
        success "Commit created successfully!"
    fi
    
    # Show the commit
    echo ""
    log "Recent commit:"
    git log --oneline -1 --color=always
    
    echo ""
    success "🎉 Commit process completed successfully!"
    echo "Your changes have been committed with validation handled by the git pre-commit hook."
}

# Run main function
main "$@"
