#!/usr/bin/env bash
set -euo pipefail

# === 🏁 Handle optional --legend-only flag ===
if [[ "${1:-}" == "--legend-only" ]]; then
  legend_md=$(cat <<EOF
## 📊 Load Test Quality Indicators

Each metric is tagged with a quality indicator:

- 🟢 Excellent – Within optimal thresholds for high-performance APIs.
- 🟡 Acceptable – Usable, but might need improvement under production load.
- 🔴 Needs Attention – Consider investigating or optimizing this area.

| Metric                | 🟢 Excellent            | 🟡 Acceptable             | 🔴 Needs Attention     |
|-----------------------|-------------------------|---------------------------|------------------------|
| **Throughput**        | >10,000 req/sec         | 1,000–10,000 req/sec      | <1,000 req/sec         |
| **Avg Latency**       | <10ms                   | 10–100ms                  | >100ms                 |
| **95th Percentile**   | <50ms                   | 50–200ms                  | >200ms                 |
| **Failure Rate**      | 0%                      | <1%                       | ≥1%                    |
| **Checks Passed**     | 100%                    | ≥95%                      | <95%                   |
| **Soak Uptime**       | 100%                    | ≥99%                      | <99%                   |
EOF
  )
  echo "$legend_md"
  [[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && echo "$legend_md" >> "$GITHUB_STEP_SUMMARY"
  exit 0
fi

file="${1:?Missing file path}"
title="${2:?Missing title}" # e.g. "Stress GET"

# Fallback if no file
if [[ ! -f "$file" ]]; then
  echo "Missing k6 output file: $file"
  exit 1
fi

# Grep helpers
get_val() {
  grep -E "$1" "$file" | head -n1 | awk -F ':' '{print $2}' | awk '{print $1}'
}

get_duration() {
  grep -A3 "http_req_duration" "$file" | grep -E 'p\(95\)' | sed -E 's/.*p\(95\)=([0-9\.]+)(ms|s|µs).*/\1/' | head -n1
}

get_avg_duration() {
  grep -A3 "http_req_duration" "$file" | grep -E 'avg=' | sed -E 's/.*avg=([0-9\.]+)(ms|s|µs).*/\1/' | head -n1
}

get_throughput() {
  grep -E 'http_reqs.*[0-9]+(\.[0-9]+)?/s' "$file" | sed -E 's/.* ([0-9\.]+)\/s.*/\1/' | head -n1
}

get_total_requests() {
  grep -E '^.*http_reqs' "$file" | head -n1 | awk -F ':' '{print $2}' | awk '{print $1}'
}

get_failure_rate() {
  grep -A3 "http_req_failed" "$file" | grep 'rate=' | sed -E 's/.*rate=([0-9\.]+).*/\1/'
}

get_check_success() {
  grep -E 'checks_succeeded' "$file" | sed -E 's/.* ([0-9]+\.[0-9]+)%.*/\1/'
}

# Emoji tag based on thresholds
get_indicator() {
  local metric="$1"
  local value="$2"

  case "$metric" in
    throughput)
      (( $(echo "$value > 10000" | bc -l) )) && echo "🟢" && return
      (( $(echo "$value >= 1000" | bc -l) )) && echo "🟡" && return
      echo "🔴"
      ;;
    avg_latency|p95_latency)
      val=$(echo "$value" | sed 's/ms//' | tr -d 'µs')
      (( $(echo "$val < 10" | bc -l) )) && echo "🟢" && return
      (( $(echo "$val <= 100" | bc -l) )) && echo "🟡" && return
      echo "🔴"
      ;;
    failure_rate)
      (( $(echo "$value == 0" | bc -l) )) && echo "🟢" && return
      (( $(echo "$value < 1" | bc -l) )) && echo "🟡" && return
      echo "🔴"
      ;;
    checks_passed)
      (( $(echo "$value == 100" | bc -l) )) && echo "🟢" && return
      (( $(echo "$value >= 95" | bc -l) )) && echo "🟡" && return
      echo "🔴"
      ;;
  esac
}

# Determine if this is soak test
is_soak=0
if grep -qi "soak" "$file"; then
  is_soak=1
elif grep -q "sleep(2)" "$file"; then
  is_soak=1
fi

# Parse data
total_requests=$(get_total_requests)
throughput=$(get_throughput)
avg_duration=$(get_avg_duration)
p95_duration=$(get_duration)
failure_rate=$(get_failure_rate)
check_passed=$(get_check_success)
uptime=""

# Mark uptime only for soak
if [[ "$is_soak" == "1" && "$failure_rate" == "0.00" && "$check_passed" == "100.00" ]]; then
  uptime="🟢 Uptime: 100%"
fi

# Calculate indicators
throughput_emoji=$(get_indicator throughput "$throughput")
avg_emoji=$(get_indicator avg_latency "$avg_duration")
p95_emoji=$(get_indicator p95_latency "$p95_duration")
fail_emoji=$(get_indicator failure_rate "$failure_rate")
check_emoji=$(get_indicator checks_passed "$check_passed")

# Format report
summary_md=$(cat <<EOF
## $title

- 💥 Total requests: $total_requests
- 🔁 Throughput: $throughput reqs/sec $throughput_emoji
- ⏱️ Avg latency: $avg_duration $avg_emoji
- 🚀 95th percentile latency: $p95_duration $p95_emoji
- ❌ Failure rate: $failure_rate $fail_emoji
- ✅ Checks passed: $check_passed% $check_emoji
${uptime:+- $uptime}
EOF
)

echo "$summary_md"
[[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && echo "$summary_md" >> "$GITHUB_STEP_SUMMARY"
