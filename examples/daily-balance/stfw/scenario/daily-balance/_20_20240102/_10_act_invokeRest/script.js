import http from 'k6/http';
import { check } from 'k6';

// Day2 (2024-01-02) の取引。前日 (Day1) の残高に加減算される。
const host = __ENV.stfw_target_host;
const txns = [
  { account_id: 'acc-001', amount: -200, bizdate: '20240102' },
  { account_id: 'acc-002', amount: 100, bizdate: '20240102' },
];

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: { checks: ['rate==1.0'] },
};

export default function () {
  for (const tx of txns) {
    const res = http.post(`http://${host}:8080/transactions`, JSON.stringify(tx), {
      headers: { 'Content-Type': 'application/json' },
    });
    check(res, { 'status is 201': (r) => r.status === 201 });
  }
}
