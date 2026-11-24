import { describe, expect, it } from "vitest";
import { toFields } from "./forms.ts";

describe("toFields", () => {
  it("keeps single fields as strings", () => {
    expect(
      toFields([
        ["name", "a"],
        ["visibility", "public"],
      ]),
    ).toEqual({ name: "a", visibility: "public" });
  });

  it("collects repeated fields into a list", () => {
    expect(
      toFields([
        ["tags", "a"],
        ["tags", "b"],
      ]),
    ).toEqual({ tags: ["a", "b"] });
  });

  it("splits fields with a separator, dropping blanks", () => {
    expect(toFields([["tags", " a, b,,c "]], new Map([["tags", ","]]))).toEqual({
      tags: ["a", "b", "c"],
    });
  });

  it("sends an empty list for an empty field with a separator", () => {
    expect(toFields([["tags", ""]], new Map([["tags", ","]]))).toEqual({ tags: [] });
  });
});
