import { check, group } from "k6";
import http from "k6/http";

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
      exec: "TestSmoke",
      tags: { scenario: "smoke" },
      executor: "constant-vus",
      vus: 1,
      duration: "5m",
    },
  },
  thresholds: {
    "http_req_duration{scenario:smoke}": ["p(99)<1000"],
    "http_req_failed{scenario:smoke}": ["rate<0.01"],
  },
};

export function TestSmoke() {
  const params = {
    headers: { "Content-Type": "application/json" },
    timeout: "1s",
  };

  TestRateAPI(params);
  TestAllocAPI(params);
}

export function TestRateAPI(params) {
  const url = `http://${__ENV.QMS_ADDR}/api/v1/allow`;

  group("QMS API", () => {
    group(
      "should return valid response from /api/v1/allow with valid payload",
      () => {
        // Given
        const namespace = `namespace${1 + Math.floor(Math.random() * 3)}`;
        const resource = `resource${1 + Math.floor(Math.random() * 10)}`;
        const payload = JSON.stringify({
          namespace: namespace,
          resource: resource,
          tokens: 1,
        });

        // When
        const res = http.post(url, payload, params);

        // Then
        check(res, {
          "is status 200": (r) => r.status === 200,
          "qms status is 1001": (r) => r.json("status") === 1001,
          "qms msg is ok": (r) => r.json("msg") === "ok",
          "qms result.wait_time is greater or equal zero": (r) =>
            r.json("result.wait_time") >= 0,
        });
      }
    );

    group(
      "should return valid response from /api/v1/allow with invalid payload",
      () => {
        // Given
        const namespace = `unknown`;
        const resource = `unknown`;
        const payload = JSON.stringify({
          namespace: namespace,
          resource: resource,
          tokens: 1,
        });

        // When
        const res = http.post(url, payload, params);

        // Then
        check(res, {
          "is status 200": (r) => r.status === 200,
          "qms status is 1002": (r) => r.json("status") === 1002,
          "qms msg is not found": (r) => r.json("msg") === "not found",
          "qms result.wait_time is zero": (r) =>
            r.json("result.wait_time") === 0,
          "qms result.ok is false": (r) => r.json("result.ok") === false,
        });
      }
    );
  });
}

export function TestAllocAPI(params) {
  group("QMS API", () => {
    group("should return a valid response from /api/v1/view", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource${1 + Math.floor(Math.random() * 5)}`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/view`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
      });

      // When
      const res = http.post(url, payload, params);

      // Then
      check(res, {
        "is status 200": (r) => r.status === 200,
        "qms status is 1001": (r) => r.json("status") === 1001,
        "qms msg is ok": (r) => r.json("msg") === "ok",
        "qms result.allocated is greater or equal zero": (r) =>
          r.json("result.allocated") >= 0,
        "qms result.capacity is greater or equal zero": (r) =>
          r.json("result.capacity") >= 0,
        "qms result.version is greater or equal zero": (r) =>
          r.json("result.version") >= 1,
      });
    });

    group(
      "should return valid response from /api/v1/view with invalid payload",
      () => {
        // Given
        const namespace = `unknown`;
        const resource = `unknown`;
        const url = `http://${__ENV.QMS_ADDR}/api/v1/view`;
        const payload = JSON.stringify({
          namespace: namespace,
          resource: resource,
          tokens: 1,
        });

        // When
        const res = http.post(url, payload, params);

        // Then
        check(res, {
          "is status 200": (r) => r.status === 200,
          "qms status is 1002": (r) => r.json("status") === 1002,
          "qms msg is not found": (r) => r.json("msg") === "not found",
          "qms result.allocated is zero": (r) =>
            r.json("result.allocated") === 0,
          "qms result.capacity is zero": (r) => r.json("result.capacity") === 0,
          "qms result.version is zero": (r) => r.json("result.version") === 0,
        });
      }
    );

    group("should return a valid response from /api/v1/alloc", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource${1 + Math.floor(Math.random() * 5)}`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/alloc`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
        version: 0,
      });

      // When
      const res = http.post(url, payload, params);

      // Then
      check(res, {
        "is status 200": (r) => r.status === 200,
        "qms status is 1001": (r) => r.json("status") === 1001,
        "qms msg is ok": (r) => r.json("msg") === "ok",
        "qms result.remaining_tokens is greater or equal zero": (r) =>
          r.json("result.remaining_tokens") >= 0,
        "qms result.current_version is greater or equal zero": (r) =>
          r.json("result.current_version") >= 0,
      });
    });

    group(
      "should return valid response from /api/v1/alloc with invalid payload",
      () => {
        // Given
        const namespace = `unknown`;
        const resource = `unknown`;
        const url = `http://${__ENV.QMS_ADDR}/api/v1/alloc`;
        const payload = JSON.stringify({
          namespace: namespace,
          resource: resource,
          tokens: 1,
        });

        // When
        const res = http.post(url, payload, params);

        // Then
        check(res, {
          "is status 200": (r) => r.status === 200,
          "qms status is 1002": (r) => r.json("status") === 1002,
          "qms msg is not found": (r) => r.json("msg") === "not found",
          "qms result.remaining_tokens is equal zero": (r) =>
            r.json("result.remaining_tokens") === 0,
          "qms result.current_version is equal zero": (r) =>
            r.json("result.current_version") === 0,
          "qms result.ok is false": (r) => r.json("result.ok") === false,
        });
      }
    );

    group("should return a valid response from /api/v1/free", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource${1 + Math.floor(Math.random() * 5)}`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/free`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
        version: 0,
      });

      // When
      const res = http.post(url, payload, params);

      // Then
      check(res, {
        "is status 200": (r) => r.status === 200,
        "qms status is 1001": (r) => r.json("status") === 1001,
        "qms msg is ok": (r) => r.json("msg") === "ok",
        "qms result.remaining_tokens is greater or equal zero": (r) =>
          r.json("result.remaining_tokens") >= 0,
        "qms result.current_version is greater or equal zero": (r) =>
          r.json("result.current_version") >= 0,
      });
    });

    group(
      "should return valid response from /api/v1/free with invalid payload",
      () => {
        // Given
        const namespace = `unknown`;
        const resource = `unknown`;
        const url = `http://${__ENV.QMS_ADDR}/api/v1/free`;
        const payload = JSON.stringify({
          namespace: namespace,
          resource: resource,
          tokens: 1,
        });

        // When
        const res = http.post(url, payload, params);

        // Then
        check(res, {
          "is status 200": (r) => r.status === 200,
          "qms status is 1002": (r) => r.json("status") === 1002,
          "qms msg is not found": (r) => r.json("msg") === "not found",
          "qms result.remaining_tokens is equal zero": (r) =>
            r.json("result.remaining_tokens") === 0,
          "qms result.current_version is equal zero": (r) =>
            r.json("result.current_version") === 0,
          "qms result.ok is false": (r) => r.json("result.ok") === false,
        });
      }
    );
  });
}
