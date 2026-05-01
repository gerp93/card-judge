#!/usr/bin/env bash
################################################################################
# Version bump
################################################################################

set -e # exit on any command error

cd "$(dirname "$0")"

################################################################################
# check for pending git changes

git fetch --quiet

if [[ -n $(git status --porcelain) ]]; then
    echo "Exiting: there are local git changes"
    exit 1
fi

if [[ $(git rev-parse @) != $(git rev-parse @{u}) ]]; then
    echo "Exiting: there are remote git changes"
    exit 1
fi

################################################################################
# get current version

GIT_TAG_VERSION=$(git tag --list 'v*' --sort=-version:refname | head -n 1)
if [[ -z "$GIT_TAG_VERSION" ]]; then
    echo "Exiting: git tag version not found"
    exit 1
fi

CURRENT_VERSION=${GIT_TAG_VERSION:1}
REGEX_VALID_VERSION="([0-9]+)\.([0-9]+)\.([0-9]+)"
if [[ ! "$CURRENT_VERSION" =~ ^$REGEX_VALID_VERSION$ ]]; then
    echo "Exiting: current version not valid ($CURRENT_VERSION)"
    exit 1
fi

VERSION_MAJOR=${BASH_REMATCH[1]}
VERSION_MINOR=${BASH_REMATCH[2]}
VERSION_PATCH=${BASH_REMATCH[3]}

################################################################################
# get new version

case "$1" in
    major)
        VERSION_MAJOR=$((10#$VERSION_MAJOR + 1));
        VERSION_MINOR=0;
        VERSION_PATCH=0;
        ;;
    minor)
        VERSION_MINOR=$((10#$VERSION_MINOR + 1));
        VERSION_PATCH=0;
        ;;
    patch)
        VERSION_PATCH=$((10#$VERSION_PATCH + 1));
        ;;
    *)
        echo "Exiting: invalid arguments provided"
        echo "Usage: $(basename "$0") [major|minor|patch]"
        exit 1
        ;;
esac

NEW_VERSION="$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH"

################################################################################
# update versions

sed -i -E 's/^Version: '$REGEX_VALID_VERSION'$/Version: '$NEW_VERSION'/' README.md

################################################################################
# print final messages

echo "Successfully updated from $CURRENT_VERSION to $NEW_VERSION"
echo "Review pending changes, if all looks good run the following:"
echo ""
echo "git add $(pwd)/README.md"
echo "git commit -m 'v$NEW_VERSION'"
echo "git push"
echo "git tag -a v$NEW_VERSION -m 'version v$NEW_VERSION'"
echo "git push --tags"

exit 0
