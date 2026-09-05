import { describe, it, expect, beforeEach } from "vitest";
import { TextProvider } from "ink";
import { render } from "ink-testing-library";
import React from "react";
import { useStore } from "../store.js";
import { StatusBar } from "../components/StatusBar.js";
import { ResultPanel } from "../components/ResultPanel.js";
import { HelpPanel } from "../components/HelpPanel.js";

describe("StatusBar", () => {
  beforeEach(() => {
    useStore.getState().reset();
  });

  it("renders disconnected state", () => {
    const { lastFrame } = render(React.createElement(StatusBar));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("○");
  });

  it("renders connected state with address", () => {
    useStore.getState().setConnected(true);
    useStore.getState().setAddress("localhost:8080");
    const { lastFrame } = render(React.createElement(StatusBar));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("●");
    expect(frame).toContain("localhost:8080");
  });

  it("renders quality score when set", () => {
    useStore.getState().setConnected(true);
    useStore.getState().setResult(95, true, [], [], []);
    const { lastFrame } = render(React.createElement(StatusBar));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("95");
    expect(frame).toContain("PASS");
  });

  it("renders FAIL when not passed", () => {
    useStore.getState().setResult(60, false, ["build"], [], []);
    const { lastFrame } = render(React.createElement(StatusBar));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("60");
    expect(frame).toContain("FAIL");
  });
});

describe("ResultPanel", () => {
  beforeEach(() => {
    useStore.getState().reset();
  });

  it("renders passed result", () => {
    useStore.getState().setResult(100, true, [], [], ["src/main.go"]);
    const { lastFrame } = render(React.createElement(ResultPanel));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("100");
    expect(frame).toContain("PASS");
  });

  it("renders blocking issues", () => {
    useStore.getState().setResult(50, false, ["build", "test"], [], []);
    const { lastFrame } = render(React.createElement(ResultPanel));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("build");
    expect(frame).toContain("test");
    expect(frame).toContain("FAIL");
  });

  it("renders artifacts", () => {
    useStore.getState().setResult(100, true, [], [], ["output/app.go", "output/test.go"]);
    const { lastFrame } = render(React.createElement(ResultPanel));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("output/app.go");
  });
});

describe("HelpPanel", () => {
  it("renders command list", () => {
    const { lastFrame } = render(React.createElement(HelpPanel));
    const frame = lastFrame() ?? "";
    expect(frame).toContain("/run");
    expect(frame).toContain("/quick");
    expect(frame).toContain("/verify");
    expect(frame).toContain("/exit");
  });
});