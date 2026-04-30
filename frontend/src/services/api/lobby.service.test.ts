import { describe, it, expect, vi, beforeEach } from "vitest";

function createLobbyUrlBuilder() {
  const httpBase = "http://localhost:3000/api/v1".replace(/\/api\/v1$/, "");
  const wsBase = httpBase.replace(/^http/, "ws");
  return (path: string, params: Record<string, string>) => {
    const searchParams = new URLSearchParams(params).toString();
    return `${wsBase}${path}?${searchParams}`;
  };
}

describe("getLobbyWsUrl logic", () => {
  const getLobbyWsUrl = createLobbyUrlBuilder();

  it("converts HTTP base URL to WS URL for visitor", () => {
    const url = getLobbyWsUrl("/ws/lobby/visitor", {
      request_id: "req-123",
      meeting_code: "abc-defg-hijk",
    });
    expect(url).toContain("ws://localhost:3000");
    expect(url).toContain("/ws/lobby/visitor");
    expect(url).toContain("request_id=req-123");
    expect(url).toContain("meeting_code=");
  });

  it("converts HTTP base URL to WS URL for admin", () => {
    const url = getLobbyWsUrl("/ws/lobby/admin", {
      meeting_code: "abc-defg-hijk",
      token: "some-jwt-token",
    });
    expect(url).toContain("ws://localhost:3000");
    expect(url).toContain("/ws/lobby/admin");
    expect(url).toContain("token=some-jwt-token");
  });

  it("URL-encodes special characters in params", () => {
    const url = getLobbyWsUrl("/ws/lobby/visitor", {
      request_id: "req 123",
      meeting_code: "abc defg",
    });
    expect(url).toContain("req+123");
    expect(url).toContain("abc+defg");
  });

  it("handles multiple query parameters", () => {
    const url = getLobbyWsUrl("/ws/lobby/visitor", {
      request_id: "r1",
      meeting_code: "m1",
    });
    const urlObj = new URL(url);
    expect(urlObj.searchParams.get("request_id")).toBe("r1");
    expect(urlObj.searchParams.get("meeting_code")).toBe("m1");
  });

  it("handles empty params", () => {
    const url = getLobbyWsUrl("/ws/lobby/visitor", {});
    expect(url).toContain("/ws/lobby/visitor?");
  });
});
