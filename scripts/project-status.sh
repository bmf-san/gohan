#!/bin/bash

# Project Management Helper Script for Gohan v1.0 Development
# Usage: ./scripts/project-status.sh

set -e

PROJECT_URL="https://github.com/users/bmf-san/projects/3"
REPO="bmf-san/gohan"

echo "🎯 Gohan v1.0 Development Status"
echo "=================================="
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed"
    echo "   Install it from: https://cli.github.com/"
    exit 1
fi

# Check authentication
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub CLI"
    echo "   Run: gh auth login"
    exit 1
fi

echo "📊 Issue Status Summary"
echo "-----------------------"

# Count issues by status
TOTAL_ISSUES=$(gh issue list --repo $REPO --state open --json number | jq length)
echo "Total Open Issues: $TOTAL_ISSUES"

# Count by labels
FOUNDATION=$(gh issue list --repo $REPO --label "foundation" --state open --json number | jq length)
CORE=$(gh issue list --repo $REPO --label "core" --state open --json number | jq length)
FEATURES=$(gh issue list --repo $REPO --label "features" --state open --json number | jq length)
QUALITY=$(gh issue list --repo $REPO --label "quality" --state open --json number | jq length)
RELEASE=$(gh issue list --repo $REPO --label "release" --state open --json number | jq length)

echo ""
echo "📋 By Phase:"
echo "  Foundation: $FOUNDATION"
echo "  Core:       $CORE"
echo "  Features:   $FEATURES"
echo "  Quality:    $QUALITY"
echo "  Release:    $RELEASE"

echo ""
echo "🔗 Quick Links"
echo "--------------"
echo "Project Board: $PROJECT_URL"
echo "Repository:    https://github.com/$REPO"
echo "Issues:        https://github.com/$REPO/issues"
echo ""

echo "📈 Recent Activity"
echo "------------------"
gh issue list --repo $REPO --limit 5 --json number,title,createdAt,labels | \
    jq -r '.[] | "Issue #\(.number): \(.title) (\(.createdAt[:10]))"'

echo ""
echo "🚀 Next Actions"
echo "---------------"
echo "1. Review current sprint progress"
echo "2. Update issue status in project board"
echo "3. Check for blockers and dependencies"
echo "4. Plan next sprint if needed"
echo ""

# Check if any issues are ready to start
READY_ISSUES=$(gh issue list --repo $REPO --label "foundation" --state open --json number,title | jq -r '.[] | "Issue #\(.number): \(.title)"')
if [ ! -z "$READY_ISSUES" ]; then
    echo "🎯 Issues Ready to Start:"
    echo "$READY_ISSUES"
else
    echo "✅ No issues waiting to start"
fi

echo ""
echo "Run with --help for more options"
