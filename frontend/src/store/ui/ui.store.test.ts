import { describe, it, expect, beforeEach, vi } from "vitest";
import { useUIStore } from "./ui.store";

describe("useUIStore", () => {
  beforeEach(() => {
    useUIStore.setState({ toasts: [], modals: {}, isLoading: false, loadingMessage: undefined });
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("starts with default state", () => {
    const state = useUIStore.getState();
    expect(state.toasts).toEqual([]);
    expect(state.modals).toEqual({});
    expect(state.isLoading).toBe(false);
    expect(state.loadingMessage).toBeUndefined();
  });

  it("shows a toast", () => {
    useUIStore.getState().showToast("Hello", "success");
    const toasts = useUIStore.getState().toasts;
    expect(toasts).toHaveLength(1);
    expect(toasts[0].message).toBe("Hello");
    expect(toasts[0].type).toBe("success");
  });

  it("shows toast with default info type", () => {
    useUIStore.getState().showToast("Info message");
    expect(useUIStore.getState().toasts[0].type).toBe("info");
  });

  it("auto-removes toast after duration", () => {
    useUIStore.getState().showToast("Auto-dismiss", "info", 3000);
    expect(useUIStore.getState().toasts).toHaveLength(1);
    vi.advanceTimersByTime(3000);
    expect(useUIStore.getState().toasts).toHaveLength(0);
  });

  it("does not auto-remove toast when duration is 0", () => {
    useUIStore.getState().showToast("Persistent", "info", 0);
    vi.advanceTimersByTime(10000);
    expect(useUIStore.getState().toasts).toHaveLength(1);
  });

  it("hides a toast by id", () => {
    useUIStore.getState().showToast("Toast 1", "info", 0);
    useUIStore.getState().showToast("Toast 2", "info", 0);
    const id = useUIStore.getState().toasts[0].id;
    useUIStore.getState().hideToast(id);
    expect(useUIStore.getState().toasts).toHaveLength(1);
    expect(useUIStore.getState().toasts[0].message).toBe("Toast 2");
  });

  it("opens a modal", () => {
    useUIStore.getState().openModal("confirm-delete", { id: 123 });
    const modal = useUIStore.getState().modals["confirm-delete"];
    expect(modal.isOpen).toBe(true);
    expect(modal.data).toEqual({ id: 123 });
  });

  it("closes a modal", () => {
    useUIStore.getState().openModal("my-modal", {});
    useUIStore.getState().closeModal("my-modal");
    expect(useUIStore.getState().modals["my-modal"].isOpen).toBe(false);
  });

  it("sets loading state", () => {
    useUIStore.getState().setLoading(true, "Processing...");
    const state = useUIStore.getState();
    expect(state.isLoading).toBe(true);
    expect(state.loadingMessage).toBe("Processing...");
  });

  it("sets loading state without message", () => {
    useUIStore.getState().setLoading(false);
    const state = useUIStore.getState();
    expect(state.isLoading).toBe(false);
    expect(state.loadingMessage).toBeUndefined();
  });
});
