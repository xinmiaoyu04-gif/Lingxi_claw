import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: { smoke: { executor: 'constant-vus', vus: 5, duration: '30s' } },
  thresholds: { http_req_failed: ['rate<0.01'], http_req_duration: ['p(95)<1000'] },
};

const baseUrl = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function () {
  const response = http.post(`${baseUrl}/api/v1/chat`, JSON.stringify({
    message: '请用一句话解释贝叶斯公式', course: '概率论',
  }), { headers: { 'Content-Type': 'application/json' } });
  check(response, {
    'HTTP response is successful': (r) => r.status >= 200 && r.status < 300,
    'uses the API response envelope': (r) => {
      try { const body = r.json(); return body.success === true && body.data && body.error === null; } catch (_) { return false; }
    },
  });
  sleep(1);
}
