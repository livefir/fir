#!/bin/bash

# Install pre-commit hook that enforces pre-commit-check.sh validation
# This script sets up the git hook to prevent commits without validation

echo "🔧 Installing Git Pre-Commit Hook"
echo "=================================="

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ ERROR: Not in a git repository root"
    echo "Please run this script from the repository root directory"
    exit 1
fi

# Check if pre-commit-check.sh exists
if [ ! -f "./scripts/pre-commit-check.sh" ]; then
    echo "❌ ERROR: ./scripts/pre-commit-check.sh not found"
    echo "This hook requires the pre-commit validation script"
    exit 1
fi

# Create the pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Pre-commit hook to enforce pre-commit-check.sh validation
# This prevents any commits unless ./scripts/pre-commit-check.sh passes completely

set -e  # Exit on any error

echo "🔒 Git Pre-Commit Hook: Enforcing pre-commit-check.sh validation"
echo "================================================================="

# Check if pre-commit-check.sh exists
if [ ! -f "./scripts/pre-commit-check.sh" ]; then
    echo "❌ ERROR: ./scripts/pre-commit-check.sh not found"
    echo "Cannot proceed with commit without validation script"
    exit 1
fi

# Make sure the script is executable
chmod +x ./scripts/pre-commit-check.sh

echo "🚀 Running mandatory pre-commit validation..."
echo "⏳ This may take a few minutes..."
echo ""

# Run the pre-commit check and capture its exit status
if ./scripts/pre-commit-check.sh; then
    echo ""
    echo "✅ Pre-commit validation PASSED"
    echo "🎉 Commit is allowed to proceed"
    echo "================================================================="
    exit 0
else
    echo ""
    echo "❌ Pre-commit validation FAILED"
    echo "🚫 COMMIT BLOCKED - Please fix issues and try again"
    echo ""
    echo "📋 To fix issues:"
    echo "   1. Review the validation output above"
    echo "   2. Fix any failing tests or quality gates"
    echo "   3. Re-run: ./scripts/pre-commit-check.sh"
    echo "   4. Only commit after seeing '✅ All quality gates passed'"
    echo ""
    echo "📖 See migration plan CRITICAL rule:"
    echo "   'NEVER COMMIT unless ./scripts/pre-commit-check.sh passes completely'"
    echo "================================================================="
    exit 1
fi
EOF

# Make the hook executable
chmod +x .git/hooks/pre-commit

echo "✅ Pre-commit hook installed successfully"
echo ""
echo "🔒 From now on, ALL commits will be blocked unless:"
echo "   ./scripts/pre-commit-check.sh passes completely"
echo ""
echo "📋 To test the hook:"
echo "   1. Make a small change: echo '# test' >> README.md"
echo "   2. Try to commit: git add README.md && git commit -m 'test'"
echo "   3. Watch the hook run the validation automatically"
echo ""
echo "🛠️  To bypass the hook (NOT RECOMMENDED):"
echo "   git commit --no-verify -m 'message'"
echo ""
echo "🗑️  To remove the hook:"
echo "   rm .git/hooks/pre-commit"
echo "=================================="
