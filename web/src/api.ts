import * as v from "valibot";

/** An error response from the server, `{errorClass, causedBy}`. */
export class ApiError extends Error {
  readonly status: number;
  readonly errorClass: string;

  constructor(status: number, errorClass: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.errorClass = errorClass;
  }
}

const errorSchema = v.object({ errorClass: v.string(), causedBy: v.string() });

/** Sends body as JSON and asks for JSON back. Resolves to the parsed body, or undefined for 204. */
export async function request(method: string, url: string, body?: unknown): Promise<unknown> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  const response = await fetch(url, {
    method,
    headers,
    body: body === undefined ? null : JSON.stringify(body),
  });
  if (response.ok) {
    return response.status === 204 ? undefined : response.json();
  }
  throw await toApiError(response);
}

async function toApiError(response: Response): Promise<ApiError> {
  const parsed = v.safeParse(errorSchema, await response.json().catch(() => undefined));
  if (parsed.success) {
    return new ApiError(response.status, parsed.output.errorClass, parsed.output.causedBy);
  }
  return new ApiError(
    response.status,
    "UNKNOWN",
    `Request failed with status ${String(response.status)}.`,
  );
}
