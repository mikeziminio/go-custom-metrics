import http from "k6/http";
import { sleep, check } from "k6";
import { Rate } from "k6/metrics";

// Custom error rate metric
const errorRate = new Rate("errors");

export const options = {
  stages: [
    { duration: "30s", target: 10 }, // Ramp-up to 10 VUs
    { duration: "1m", target: 10 }, // Stay at 10 VUs
    { duration: "30s", target: 0 }, // Ramp-down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    errors: ["rate<0.01"], // Error rate should be less than 1%
  },
};

const BASE_URL = "http://localhost:8080";

// Helper function to generate random metric names
function getRandomMetricName() {
  const chars = "abcdefghijklmnopqrstuvwxyz";
  let result = "";
  for (let i = 0; i < 10; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

// Helper function to generate random metric values
function getRandomGaugeValue() {
  return Math.random() * 1000;
}

function getRandomCounterValue() {
  return Math.floor(Math.random() * 1000);
}

export default function () {
  // Test 1: POST /update - Update a single gauge metric
  const gaugeUpdatePayload = {
    id: `${getRandomMetricName()}_gauge`,
    type: "gauge",
    value: getRandomGaugeValue(),
  };

  let res = http.post(
    `${BASE_URL}/update`,
    JSON.stringify(gaugeUpdatePayload),
    {
      headers: { "Content-Type": "application/json" },
    },
  );

  check(res, {
    "POST /update gauge status is 200": (r) => r.status === 200,
    "POST /update gauge response has id": (r) => r.json().id !== undefined,
    "POST /update gauge response has type": (r) => r.json().type !== undefined,
    "POST /update gauge response has value": (r) =>
      r.json().value !== undefined,
  });

  errorRate.add(res.status !== 200);

  // Test 2: POST /update - Update a single counter metric
  const counterUpdatePayload = {
    id: `${getRandomMetricName()}_counter`,
    type: "counter",
    delta: getRandomCounterValue(),
  };

  res = http.post(`${BASE_URL}/update`, JSON.stringify(counterUpdatePayload), {
    headers: { "Content-Type": "application/json" },
  });

  check(res, {
    "POST /update counter status is 200": (r) => r.status === 200,
    "POST /update counter response has id": (r) => r.json().id !== undefined,
    "POST /update counter response has type": (r) =>
      r.json().type !== undefined,
    "POST /update counter response has delta": (r) =>
      r.json().delta !== undefined,
  });

  errorRate.add(res.status !== 200);

  // Test 3: POST /update/{metricType}/{metricName}/{value} - Direct URL parameter update
  const metricName = `${getRandomMetricName()}_direct`;
  const gaugeValue = getRandomGaugeValue();

  res = http.post(`${BASE_URL}/update/gauge/${metricName}/${gaugeValue}`);

  check(res, {
    "POST /update/{metricType}/{metricName}/{value} gauge status is 200": (r) =>
      r.status === 200,
  });

  errorRate.add(res.status !== 200);

  // Test 4: POST /updates - Batch update multiple metrics
  const updatesPayload = [
    {
      id: `${getRandomMetricName()}_batch1`,
      type: "gauge",
      value: getRandomGaugeValue(),
    },
    {
      id: `${getRandomMetricName()}_batch2`,
      type: "counter",
      delta: getRandomCounterValue(),
    },
  ];

  res = http.post(`${BASE_URL}/updates`, JSON.stringify(updatesPayload), {
    headers: { "Content-Type": "application/json" },
  });

  check(res, {
    "POST /updates status is 200": (r) => r.status === 200,
  });

  errorRate.add(res.status !== 200);

  // Test 5: POST /value - Get specific metric by type and name
  const getPayload = {
    id: `${getRandomMetricName()}_get`,
    type: "gauge",
  };

  // First update it to make sure it exists
  http.post(
    `${BASE_URL}/update`,
    JSON.stringify({
      id: getPayload.id,
      type: getPayload.type,
      value: getRandomGaugeValue(),
    }),
    {
      headers: { "Content-Type": "application/json" },
    },
  );

  // Then get it
  res = http.post(`${BASE_URL}/value`, JSON.stringify(getPayload), {
    headers: { "Content-Type": "application/json" },
  });

  check(res, {
    "POST /value status is 200": (r) => r.status === 200,
    "POST /value response has id": (r) => r.json().id !== undefined,
    "POST /value response has type": (r) => r.json().type !== undefined,
  });

  errorRate.add(res.status !== 200);

  // Test 6: GET /value/{metricType}/{metricName} - Get metric by URL parameters
  const metricToGet = `${getRandomMetricName()}_by_param`;
  const getValue = getRandomGaugeValue();

  // First update it
  http.post(`${BASE_URL}/update/gauge/${metricToGet}/${getValue}`);

  // Then get it via URL params
  res = http.get(`${BASE_URL}/value/gauge/${metricToGet}`);

  check(res, {
    "GET /value/{metricType}/{metricName} status is 200": (r) =>
      r.status === 200,
    "GET /value/{metricType}/{metricName} response is not empty": (r) =>
      r.body !== "",
  });

  errorRate.add(res.status !== 200);

  // Test 7: GET /ping - Health check endpoint
  res = http.get(`${BASE_URL}/ping`);

  check(res, {
    "GET /ping status is 200": (r) => r.status === 200,
  });

  errorRate.add(res.status !== 200);

  // Test 8: GET / - List all metrics
  res = http.get(`${BASE_URL}/`);

  check(res, {
    "GET / status is 200": (r) => r.status === 200,
  });

  errorRate.add(res.status !== 200);

  // Test 9: POST /update with invalid metric type (should return 400)
  const invalidPayload = {
    id: `${getRandomMetricName()}_invalid`,
    type: "invalid_type",
    value: 123.45,
  };

  res = http.post(`${BASE_URL}/update`, JSON.stringify(invalidPayload), {
    headers: { "Content-Type": "application/json" },
  });

  check(res, {
    "POST /update invalid type returns 400": (r) => r.status === 400,
  });

  errorRate.add(res.status !== 400);

  // Test 10: GET /value/{metricType}/{metricName} for non-existent metric (should return 404)
  const nonExistentMetric = `${getRandomMetricName()}_nonexistent`;

  res = http.get(`${BASE_URL}/value/gauge/${nonExistentMetric}`);

  check(res, {
    "GET /value/{metricType}/{metricName} non-existent returns 404": (r) =>
      r.status === 404,
  });

  errorRate.add(res.status !== 404);

  sleep(1);
}
