#!/usr/bin/env bash

COVERAGE_FILE="coverage.out"
FILTERED_FILE="coverage.filtered.out"
MIN_COVERAGE=70
IGNORE_FILE=".coverignore"

if [ ! -f "$COVERAGE_FILE" ]; then
  echo "❌ Coverage file not found: $COVERAGE_FILE"
  exit 1
fi

# Build ignore regex
IGNORE_PATTERN=""
IGNORED_COUNT=0
if [ -f "$IGNORE_FILE" ]; then
  while read -r line; do
    [ -z "$line" ] && continue
    escaped=$(echo "$line" | sed 's/[].[^$*\/]/\\&/g')
    pattern="${escaped}$"
    if [ -z "$IGNORE_PATTERN" ]; then
      IGNORE_PATTERN="$pattern"
    else
      IGNORE_PATTERN="$IGNORE_PATTERN|$pattern"
    fi
  done < "$IGNORE_FILE"
fi

# Filter out ignored files
echo "mode: set" > "$FILTERED_FILE"
grep -v "^mode: " "$COVERAGE_FILE" | {
  while read -r line; do
    file=$(echo "$line" | cut -d ':' -f1)
    # Strip module prefix
    stripped=${file#github.com*/}
    if [[ -n "$IGNORE_PATTERN" && "$stripped" =~ $IGNORE_PATTERN ]]; then
      ((IGNORED_COUNT++))
      continue
    fi
    echo "$line" >> "$FILTERED_FILE"
  done
}

if [ "$IGNORED_COUNT" -gt 0 ]; then
  echo "🛑 Ignored $IGNORED_COUNT file(s) from coverage based on $IGNORE_FILE"
fi

# Per-directory coverage
echo -e "\n📁 Code Coverage per Directory:\n"

go tool cover -func="$FILTERED_FILE" | grep -vE 'total:' | awk '
{
  split($1, parts, ":")
  filepath = parts[1]
  sub(/^github.com\/[^\/]+\/[^\/]+\//, "", filepath)
  n = split(filepath, pathparts, "/")

  dir = ""
  for (i = 1; i < n; i++) {
    dir = (dir == "" ? pathparts[i] : dir "/" pathparts[i])
  }

  if ($3 ~ /[0-9]+%$/) {
    sub(/%$/, "", $3)
    count[dir]++
    sum[dir] += $3
  }
}
END {
  for (d in sum) {
    printf "  %-40s %5.1f%%\n", d, sum[d] / count[d]
  }
}' | sort

# Total coverage
total=$(go tool cover -func="$FILTERED_FILE" | grep total: | awk '{gsub(/%/, ""); print $3}')
echo -e "\n🎯 Total Code Coverage: ${total}%"

if (( $(echo "$total < $MIN_COVERAGE" | bc -l) )); then
  echo "❌ Code coverage ${total}% is below required minimum of ${MIN_COVERAGE}%"
  exit 1
else
  echo "✅ Code coverage check passed"
fi

# Cleanup
rm -f "$FILTERED_FILE"
