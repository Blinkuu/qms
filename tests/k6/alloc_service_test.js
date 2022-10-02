import {Counter} from "k6/metrics";
import http from "k6/http";

const requestTotalCounter = new Counter("request_total")
const requestTotalFailureCounter = new Counter("request_total_failure")

export default function () {
    alloc(1)
    free(1)
}

function alloc(tokens) {
    const url = `http://${__ENV.ADDR}/api/v1/alloc`;
    const payload = JSON.stringify({
        namespace: "namespace1",
        resource: "resource1",
        tokens: tokens,
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
    }
}

function free(tokens) {
    const url = `http://${__ENV.ADDR}/api/v1/free`;
    const payload = JSON.stringify({
        namespace: "namespace1",
        resource: "resource1",
        tokens: tokens,
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
    }
}
