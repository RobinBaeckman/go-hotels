import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  vus: 10,
  duration: '15s',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
    checks: ['rate==1.0'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)'], 
};

export default function () {
  const res = http.get('http://localhost:8080/ready');
  check(res, {
    'service is ready': (r) => r.status === 200,
  });
  sleep(2);
}
