import {Counter} from "k6/metrics";
import http from "k6/http";

const requestTotalCounter = new Counter("request_total")
const requestTotalFailureCounter = new Counter("request_total_failure")

export default function () {
    const url = `http://${__ENV.ADDR}/api/v1/allow`;
    const payload = JSON.stringify({
        namespace: "namespace1",
        resource: "resource1",
        tokens: 1,
    });
    const params = {
        headers: {
            "Content-Type": "application/json",
        },
    };

    requestTotalCounter.add(1)
    const res = http.post(url, payload, params);
    if (res.status != 200) {
        requestTotalFailureCounter.add(1)
        return;
    }

    const parsedRes = res.json();
    if (parsedRes.status == 1001) {
        if (parsedRes.result.ok && parsedRes.result.wait_time == 0) {
            hit_counter.add(1);
            const targetUrl = "http://localhost:8080/api/v1/ping";
            http.get(targetUrl)
        }
    }
}
