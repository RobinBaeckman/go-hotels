import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '5s', target: 50 },
    { duration: '5s', target: 100 },
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
  const res = http.get('http://localhost:8080/hotels?city=Tokyo');
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
