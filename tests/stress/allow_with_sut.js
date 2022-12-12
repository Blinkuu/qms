import { Counter } from "k6/metrics";
import http from "k6/http";
import { sleep } from "k6";

export const options = {
  ext: {
    loadimpact: {
      projectID: 3593275,
      name: "Test QMS",
      distribution: {
        frankfurtDistribution: {
          loadZone: "amazon:de:frankfurt",
          percent: 100,
        },
      },
    },
  },
  scenarios: {
    stress: {
      exec: "TestAllowAPI",
      executor: "ramping-vus",
      tags: { scenario: "stress" },
      stages: [
        { target: 100, duration: "30s" },
        { target: 100, duration: "3m" },
        { target: 150, duration: "30s" },
        { target: 150, duration: "3m" },
        { target: 200, duration: "30s" },
        { target: 200, duration: "3m" },
        { target: 250, duration: "30s" },
        { target: 250, duration: "3m" },
        { target: 300, duration: "30s" },
        { target: 300, duration: "3m" },
        { target: 0, duration: "30s" },
      ],
      gracefulRampDown: "30s",
      gracefulStop: "30s",
    },
  },
  thresholds: {
    // Stress
    "http_req_duration{scenario:stress}": ["p(99)<1000"],
    "http_req_failed{scenario:stress}": ["rate<0.01"],
  },
};

export function TestAllowAPI() {
  const namespace = `namespace1`;
  const resource = `resource1`;
  const url = `http://${__ENV.QMS_ADDR}/api/v1/allow`;
  const payload = JSON.stringify({
    namespace: namespace,
    resource: resource,
    tokens: 1,
  });
  const params = { headers: { "Content-Type": "application/json" } };

  const res = http.post(url, payload, params);
  if (res.status != 200) {
    return;
  }

  const parsedRes = res.json();
  if (parsedRes.status == 1001) {
    if (parsedRes.result.ok && parsedRes.result.wait_time == 0) {
      const targetUrl = `http://${__ENV.SUT_ADDR}/api/v1/ping`;
      http.get(targetUrl);
    }
  }

  sleep(0.2);
}
