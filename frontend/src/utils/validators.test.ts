import { describe, it, expect } from "vitest";
import { isValidMeetingCode } from "./validators";

describe("isValidMeetingCode", () => {
  it("accepts valid meeting code xxx-xxxx-xxx", () => {
    expect(isValidMeetingCode("abc-defg-hijk")).toBe(true);
  });

  it("accepts xx-xxxx-xx variant (11 chars)", () => {
    expect(isValidMeetingCode("ab-cdef-gh")).toBe(true);
  });

  it("rejects empty string", () => {
    expect(isValidMeetingCode("")).toBe(false);
  });

  it("rejects code without hyphens", () => {
    expect(isValidMeetingCode("abcdefghijk")).toBe(false);
  });

  it("rejects code starting with hyphen", () => {
    expect(isValidMeetingCode("-abc-defg")).toBe(false);
  });

  it("rejects code ending with hyphen", () => {
    expect(isValidMeetingCode("abc-defg-")).toBe(false);
  });

  it("rejects code with consecutive hyphens", () => {
    expect(isValidMeetingCode("abc--defg")).toBe(false);
  });

  it("rejects code with uppercase letters", () => {
    expect(isValidMeetingCode("ABC-DEFG-HIJK")).toBe(false);
  });

  it("rejects code with numbers", () => {
    expect(isValidMeetingCode("abc-1234-hijk")).toBe(false);
  });

  it("rejects code that is too short (< 8 chars)", () => {
    expect(isValidMeetingCode("a-b")).toBe(false);
  });

  it("rejects code that is too long (> 15 chars)", () => {
    expect(isValidMeetingCode("abcdefghijklmno")).toBe(false);
  });

  it("rejects code with more than 3 hyphens", () => {
    expect(isValidMeetingCode("a-b-c-d")).toBe(false);
  });

  it("accepts code with exactly 1 hyphen", () => {
    expect(isValidMeetingCode("abc-defg")).toBe(true);
  });

  it("rejects code with spaces", () => {
    expect(isValidMeetingCode("abc defg hijk")).toBe(false);
  });

  it("rejects code with special characters", () => {
    expect(isValidMeetingCode("abc_defg_hijk")).toBe(false);
  });
});
