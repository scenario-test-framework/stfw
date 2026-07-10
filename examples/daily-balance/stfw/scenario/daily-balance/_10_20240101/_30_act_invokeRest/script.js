import http from 'k6/http';
import { check } from 'k6';

// Day1 (2024-01-01) の取引。invokeRest が host_group=api の先頭ホストを
// __ENV.stfw_target_host に注入する。
const host = __ENV.stfw_target_host;
const txns = [
  { account_id: 'acc-001', amount: 500, bizdate: '20240101' },
  { account_id: 'acc-002', amount: 300, bizdate: '20240101' },
];

export const options = {
  vus: 1,
  iterations: 1,
  // 全リクエストが 201 でなければ閾値失敗 → k6 が非 0 終了 → Act 失敗 (exit 6)
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
