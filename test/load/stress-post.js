import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '5s', target: 20 },
    { duration: '5s', target: 40 },
    { duration: '5s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<800'],
    http_req_failed: ['rate<0.02'],
    checks: ['rate>0.95'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)'],
};

export default function () {
  const payload = JSON.stringify({
    name: 'Stress Hotel',
    city: 'Tokyo',
    stars: 4,
    price_per_night: 100.0,
    amenities: ['wifi', 'tv'],
  });

  const headers = { 'Content-Type': 'application/json' };
  const res = http.post('http://localhost:8080/hotels', payload, { headers });

  check(res, {
    'hotel created': (r) => r.status === 201,
  });
}
