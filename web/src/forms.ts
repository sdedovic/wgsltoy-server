import { ApiError, request } from "./api.ts";

/**
 * HTML forms only send GET and POST, so a `<form data-method="PUT">` is sent with fetch instead. Its
 * fields go as JSON, or in the query string for DELETE, and once the server accepts them the page goes
 * to `data-redirect`, or reloads. A field with `data-split=","` is sent as a list split on that.
 * `data-confirm` asks before sending.
 */
export function enhanceForms(): void {
  for (const form of document.querySelectorAll<HTMLFormElement>("form[data-method]")) {
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      void submit(form);
    });
  }
}

async function submit(form: HTMLFormElement): Promise<void> {
  const { method = "POST", confirm: question, redirect } = form.dataset;
  if (question !== undefined && !window.confirm(question)) {
    return;
  }

  const url = new URL(
    form.getAttribute("action") ?? window.location.pathname,
    window.location.href,
  );
  const fields = toFields(formEntries(form), separators(form));
  const buttons = form.querySelectorAll<HTMLButtonElement>("button");
  buttons.forEach((b) => (b.disabled = true));
  try {
    if (method === "DELETE") {
      for (const [name, value] of Object.entries(fields)) {
        for (const item of [value].flat()) {
          url.searchParams.append(name, item);
        }
      }
      await request(method, url.toString());
    } else {
      await request(method, url.toString(), fields);
    }
    if (redirect === undefined) {
      window.location.reload();
    } else {
      window.location.assign(redirect);
    }
  } catch (error) {
    showError(form, error instanceof ApiError ? error.message : "Something went wrong, try again.");
    buttons.forEach((b) => (b.disabled = false));
  }
}

// FormData includes fields outside the form that point at it with form="id"
function formEntries(form: HTMLFormElement): [string, string][] {
  return [...new FormData(form)].flatMap(([name, value]) =>
    typeof value === "string" ? [[name, value] as [string, string]] : [],
  );
}

function separators(form: HTMLFormElement): Map<string, string> {
  const result = new Map<string, string>();
  for (const element of form.elements) {
    if (
      element instanceof HTMLElement &&
      "name" in element &&
      element.dataset["split"] !== undefined
    ) {
      result.set(String(element.name), element.dataset["split"]);
    }
  }
  return result;
}

/** Form entries as an object. A name given more than once, or with a separator, is a list. */
export function toFields(
  entries: Iterable<[string, string]>,
  separators: ReadonlyMap<string, string> = new Map(),
): Record<string, string | string[]> {
  const fields: Record<string, string | string[]> = {};
  for (const [name, value] of entries) {
    const separator = separators.get(name);
    if (separator !== undefined) {
      const items = value
        .split(separator)
        .map((item) => item.trim())
        .filter((item) => item !== "");
      fields[name] = [...[fields[name] ?? []].flat(), ...items];
    } else if (name in fields) {
      fields[name] = [...[fields[name] ?? []].flat(), value];
    } else {
      fields[name] = value;
    }
  }
  return fields;
}

function showError(form: HTMLFormElement, message: string): void {
  let alert = form.querySelector<HTMLElement>(":scope > [role=alert]");
  if (alert === null) {
    alert = document.createElement("p");
    alert.className = "error";
    alert.setAttribute("role", "alert");
    form.prepend(alert);
  }
  alert.textContent = message;
}
