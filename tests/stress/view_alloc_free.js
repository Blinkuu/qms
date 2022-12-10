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
      exec: "TestViewAllocFreeAPI",
      executor: "ramping-vus",
      tags: { scenario: "stress" },
      stages: [
        // { target: 100, duration: "30s" },
        // { target: 100, duration: "3m" },
        // { target: 150, duration: "30s" },
        // { target: 150, duration: "3m" },
        // { target: 200, duration: "30s" },
        // { target: 200, duration: "3m" },
        // { target: 250, duration: "30s" },
        // { target: 250, duration: "3m" },
        // { target: 300, duration: "30s" },
        // { target: 300, duration: "3m" },
        // { target: 350, duration: "30s" },
        // { target: 350, duration: "3m" },
        // { target: 400, duration: "30s" },
        // { target: 400, duration: "3m" },
        // { target: 500, duration: "30s" },
        // { target: 500, duration: "3m" },
        { target: 50, duration: "30s" },
        { target: 50, duration: "3m" },
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

export function TestViewAllocFreeAPI() {
  TestViewAPI();
  TestViewAPI();
  TestAllocAPI()
  TestFreeAPI()

  sleep(0.8);
}

function TestViewAPI() {
    const namespace = `namespace1`;
    const resource = `resource${1 + Math.floor(Math.random() * 5)}`;
    const url = `http://${__ENV.QMS_ADDR}/api/v1/view`;
    const payload = JSON.stringify({
      namespace: namespace,
      resource: resource,
    });
    const params = { headers: { "Content-Type": "application/json" } };
  
    http.post(url, payload, params);
}

function TestAllocAPI() {
  const namespace = `namespace1`;
  const resource = `resource${1 + Math.floor(Math.random() * 5)}`;
  const url = `http://${__ENV.QMS_ADDR}/api/v1/alloc`;
  const payload = JSON.stringify({
    namespace: namespace,
    resource: resource,
    tokens: 1,
    version: 0,
  });
  const params = { headers: { "Content-Type": "application/json" } };

  http.post(url, payload, params);
}

function TestFreeAPI() {
  const namespace = `namespace1`;
  const resource = `resource${1 + Math.floor(Math.random() * 5)}`;
  const url = `http://${__ENV.QMS_ADDR}/api/v1/free`;
  const payload = JSON.stringify({
    namespace: namespace,
    resource: resource,
    tokens: 1,
    version: 0,
  });
  const params = { headers: { "Content-Type": "application/json" } };

  http.post(url, payload, params);
}

