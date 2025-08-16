import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// --- Config (works with k6 v1.1.0) ---
export const options = {
  scenarios: {
    traffic: {
      executor: 'constant-arrival-rate',   // open model
      rate: 30,                             // 30 iterations per second (~30 RPS)
      timeUnit: '1s',
      duration: '12h',                      // must be > 0
      preAllocatedVUs: 20,                  // initial VUs to sustain the rate
      maxVUs: 100,                          // headroom for spikes
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],        // keep failures under 5%
    http_req_duration: ['p(95)<800'],      // p95 < 800ms
  },
  summaryTrendStats: ['avg','min','med','max','p(90)','p(95)'],
};

// Weighted random: ~70% GET /hotels, ~25% POST /hotels, ~5% GET /ready
function pick() {
  const r = Math.random();
  if (r < 0.70) return 'GET_HOTELS';
  if (r < 0.95) return 'POST_HOTELS';
  return 'READY';
}

export default function () {
  const op = pick();

  if (op === 'GET_HOTELS') {
    const res = http.get(`${BASE_URL}/hotels?city=Tokyo`);
    check(res, { 'GET /hotels 200': (r) => r.status === 200 });
  } else if (op === 'POST_HOTELS') {
    const payload = JSON.stringify({
      name: `Hotel ${Math.random().toString(16).slice(2, 8)}`,
      city: 'Tokyo',
      stars: 4,
      price_per_night: 100.0,
      amenities: ['wifi', 'tv'],
    });
    const headers = { 'Content-Type': 'application/json' };
    const res = http.post(`${BASE_URL}/hotels`, payload, { headers });
    check(res, { 'POST /hotels 201': (r) => r.status === 201 });
  } else {
    const res = http.get(`${BASE_URL}/ready`);
    check(res, { 'GET /ready 200': (r) => r.status === 200 });
  }

  // tiny jitter so calls don’t align too perfectly
  sleep(Math.random() * 0.2);
}
