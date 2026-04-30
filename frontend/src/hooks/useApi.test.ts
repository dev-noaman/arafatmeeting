import { describe, it, expect, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useApi } from "./useApi";

describe("useApi", () => {
  let mockFn: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFn = vi.fn();
  });

  it("starts with null data and no loading", () => {
    const { result } = renderHook(() => useApi(mockFn));
    expect(result.current.data).toBeNull();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("sets loading and data on success", async () => {
    mockFn.mockResolvedValue({ id: 1, name: "Test" });
    const { result } = renderHook(() => useApi(mockFn));

    await act(async () => {
      await result.current.execute();
    });

    expect(result.current.data).toEqual({ id: 1, name: "Test" });
    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("sets error on failure", async () => {
    mockFn.mockRejectedValue({
      response: { data: { message: "Something went wrong" } },
      message: "Request failed",
    });
    const { result } = renderHook(() => useApi(mockFn));

    await act(async () => {
      try {
        await result.current.execute();
      } catch {
        // expected
      }
    });

    expect(result.current.error).toBe("Something went wrong");
    expect(result.current.isLoading).toBe(false);
  });

  it("passes arguments to the API function", async () => {
    mockFn.mockResolvedValue("ok");
    const { result } = renderHook(() =>
      useApi<(id: number, name: string) => Promise<string>, [number, string]>(mockFn)
    );

    await act(async () => {
      await result.current.execute(42, "hello");
    });

    expect(mockFn).toHaveBeenCalledWith(42, "hello");
  });

  it("resets state", async () => {
    mockFn.mockResolvedValue({ id: 1 });
    const { result } = renderHook(() => useApi(mockFn));

    await act(async () => {
      await result.current.execute();
    });
    expect(result.current.data).not.toBeNull();

    act(() => {
      result.current.reset();
    });
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });
});
