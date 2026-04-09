import http from "k6/http";
import { check, fail } from "k6";

const baseURL = __ENV.BASE_URL || "http://localhost:8081";
const sharedURL = __ENV.SHARED_URL || "https://example.com/load-test/shared";
const vus = Number(__ENV.VUS || 200);
const duration = __ENV.DURATION || "15s";

export const options = {
  vus,
  duration,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:post_links}": ["p(99)<500"],
    "http_req_duration{endpoint:get_link}": ["p(99)<500"],
  },
};

function parseJSON(res, label) {
  try {
    return res.json();
  } catch (err) {
    fail(`${label}: invalid JSON response`);
  }
}

export function setup() {
  const payload = JSON.stringify({ url: sharedURL });
  const params = {
    headers: { "Content-Type": "application/json" },
    tags: { endpoint: "post_links" },
  };

  const res = http.post(`${baseURL}/links`, payload, params);
  const ok = check(res, {
    "setup create status is 200 or 201": (r) =>
      r.status === 200 || r.status === 201,
  });
  if (!ok) {
    fail(`setup create failed with status ${res.status}`);
  }

  const body = parseJSON(res, "setup create");
  const code = body && body.code;
  if (typeof code !== "string" || code.length !== 10) {
    fail("setup create: invalid short code");
  }

  return { code };
}

export default function (data) {
  const payload = JSON.stringify({ url: sharedURL });
  const postParams = {
    headers: { "Content-Type": "application/json" },
    tags: { endpoint: "post_links" },
  };

  const createRes = http.post(`${baseURL}/links`, payload, postParams);
  const createBody = parseJSON(createRes, "create");

  check(createRes, {
    "create status is 200 or 201": (r) => r.status === 200 || r.status === 201,
  });

  check(createBody, {
    "create returns expected code": (body) => body && body.code === data.code,
  });

  const getRes = http.get(`${baseURL}/links/${data.code}`, {
    tags: { endpoint: "get_link" },
  });

  const getBody = parseJSON(getRes, "resolve");
  check(getRes, {
    "get status is 200": (r) => r.status === 200,
  });
  check(getBody, {
    "get returns expected url": (body) => body && body.url === sharedURL,
  });
}
