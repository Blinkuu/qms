import {
  describe,
  expect,
} from "https://jslib.k6.io/k6chaijs/4.3.4.2/index.js";
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
  TestRateAPI();
  TestAllocAPI();
}

export function TestRateAPI() {
  const url = `http://${__ENV.QMS_ADDR}/api/v1/allow`;

  describe("QMS API", () => {
    describe("should return valid response from /api/v1/allow with valid payload", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource1`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const expected = {
        status: 1001,
        msg: "ok",
        result: {
          wait_time: 0,
          ok: true,
        },
      };

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(JSON.parse(res.body)).to.deep.equal(expected);
    });

    describe("should return valid response from /api/v1/allow with invalid payload", () => {
      // Given
      const namespace = `unknown`;
      const resource = `unknown`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const expected = {
        status: 1002,
        msg: "not found",
        result: {
          wait_time: 0,
          ok: false,
        },
      };

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(JSON.parse(res.body)).to.deep.equal(expected);
    });
  });
}

export function TestAllocAPI() {
  describe("QMS API", () => {
    describe("should return a valid response from /api/v1/view", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource1`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/view`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const viewRes = JSON.parse(res.body);

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(viewRes.status).to.equal(1001);
      expect(viewRes.msg).to.equal("ok");
      expect(viewRes.result.allocated).to.be.at.least(0);
      expect(viewRes.result.capacity).to.be.at.least(0);
      expect(viewRes.result.version).to.be.at.least(1);
    });

    describe("should return valid response from /api/v1/view with invalid payload", () => {
      // Given
      const namespace = `unknown`;
      const resource = `unknown`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/view`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const expected = {
        status: 1002,
        msg: "not found",
        result: {
          allocated: 0,
          capacity: 0,
          version: 0,
        },
      };

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(JSON.parse(res.body)).to.deep.equal(expected);
    });

    describe("should return a valid response from /api/v1/alloc", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource1`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/alloc`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
        version: 0,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const allocRes = JSON.parse(res.body);

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(allocRes.status).to.equal(1001);
      expect(allocRes.msg).to.equal("ok");
      expect(allocRes.result.remaining_tokens).to.be.at.least(0);
      expect(allocRes.result.current_version).to.be.at.least(0);
      expect(allocRes.result.ok).to.be.true;
    });

    describe("should return valid response from /api/v1/alloc with invalid payload", () => {
      // Given
      const namespace = `unknown`;
      const resource = `unknown`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/alloc`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const expected = {
        status: 1002,
        msg: "not found",
        result: {
          remaining_tokens: 0,
          current_version: 0,
          ok: false,
        },
      };

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(JSON.parse(res.body)).to.deep.equal(expected);
    });

    describe("should return a valid response from /api/v1/free", () => {
      // Given
      const namespace = `namespace1`;
      const resource = `resource1`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/free`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
        version: 0,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const freeRes = JSON.parse(res.body);

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(freeRes.status).to.equal(1001);
      expect(freeRes.msg).to.equal("ok");
      expect(freeRes.result.remaining_tokens).to.be.at.least(0);
      expect(freeRes.result.current_version).to.be.at.least(0);
      expect(freeRes.result.ok).to.be.true;
    });

    describe("should return valid response from /api/v1/free with invalid payload", () => {
      // Given
      const namespace = `unknown`;
      const resource = `unknown`;
      const url = `http://${__ENV.QMS_ADDR}/api/v1/free`;
      const payload = JSON.stringify({
        namespace: namespace,
        resource: resource,
        tokens: 1,
      });
      const params = { headers: { "Content-Type": "application/json" } };

      // When
      const res = http.post(url, payload, params);

      // Then
      const expected = {
        status: 1002,
        msg: "not found",
        result: {
          remaining_tokens: 0,
          current_version: 0,
          ok: false,
        },
      };

      expect(res).to.have.validJsonBody();
      expect(res.status).to.equal(200);
      expect(JSON.parse(res.body)).to.deep.equal(expected);
    });
  });
}
