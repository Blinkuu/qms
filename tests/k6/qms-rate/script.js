import { Counter } from "k6/metrics";
import http from "k6/http";

const requestTotalCounter = new Counter("request_total");
const requestTotalFailureCounter = new Counter("request_total_failure");
const requestAllowedTotalCounter = new Counter("request_allowed_total");

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
        smoke: {
            exec: "FunctionalTestAllow",
            executor: "constant-vus",
            tags: { scenario: "smoke" },
            vus: 1,
            duration: "30s",
        },
        stress: {
            exec: "StressTestAllow",
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
                { target: 350, duration: "30s" },
                { target: 350, duration: "3m" },
                { target: 400, duration: "30s" },
                { target: 400, duration: "3m" },
                { target: 0, duration: "30s" },
            ],
            startTime: "30s", // Start after smoke scenario
            gracefulRampDown: "30s",
            gracefulStop: "30s",
        },
    },
    thresholds: {
        // Smoke
        "http_req_duration{scenario:smoke}": ["p(99)<1000"],
        "http_req_failed{scenario:smoke}": ["rate<0.01"],

        // Stress
        "http_req_duration{scenario:stress}": ["p(99)<1000"],
        "http_req_failed{scenario:stress}": ["rate<0.01"],
    },
};

export function FunctionalTestAllow() {
    const namespace = `namespace1`;
    const resource = `resource1`;

    const url = `http://${__ENV.QMS_ADDR}/api/v1/allow`;
    const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
    });
    const params = {
        headers: {
            "Content-Type": "application/json",
        },
    };

    requestTotalCounter.add(1);
    const res = http.post(url, payload, params);
    if (res.status != 200) {
        requestTotalFailureCounter.add(1);
        return;
    }

    const parsedRes = res.json();
    if (parsedRes.status == 1001) {
        if (parsedRes.result.ok && parsedRes.result.wait_time == 0) {
            requestAllowedTotalCounter.add(1);
            const targetUrl = `http://${__ENV.SUT_ADDR}/api/v1/ping`;
            http.get(targetUrl);
        }
    }
}

export function StressTestAllow() {
    const namespace = `namespace${Math.ceil(Math.random() * 3)}`;
    const resource = `resource${Math.ceil(Math.random() * 10)}`;

    const url = `http://${__ENV.QMS_ADDR}/api/v1/allow`;
    const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
    });
    const params = {
        headers: {
            "Content-Type": "application/json",
        },
    };

    requestTotalCounter.add(1);
    const res = http.post(url, payload, params);
    if (res.status != 200) {
        requestTotalFailureCounter.add(1);
        return;
    }
}
