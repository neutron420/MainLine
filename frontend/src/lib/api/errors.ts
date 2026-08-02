import { Code, ConnectError } from "@connectrpc/connect";

export function isConnectError(err: unknown): err is ConnectError {
  return err instanceof ConnectError;
}

export function isUnauthenticated(err: unknown): boolean {
  return err instanceof ConnectError && err.code === Code.Unauthenticated;
}

export function isNotFound(err: unknown): boolean {
  return err instanceof ConnectError && err.code === Code.NotFound;
}

export function getApiErrorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.Unauthenticated:
        return "Session expired. Please log in again.";
      case Code.NotFound:
        return "The requested resource was not found.";
      case Code.AlreadyExists:
        return "That resource already exists.";
      case Code.InvalidArgument:
        return err.rawMessage || "Invalid input provided.";
      case Code.FailedPrecondition:
        return err.rawMessage || "This operation is not allowed right now.";
      case Code.PermissionDenied:
        return "You do not have permission to do this.";
      case Code.ResourceExhausted:
        return "Rate limit exceeded. Please try again in a moment.";
      case Code.DeadlineExceeded:
        return "The request timed out. Please try again.";
      case Code.Unavailable:
        return "The backend is unreachable. Please make sure the API is running.";
      case Code.Internal:
        return "Something went wrong on the server. Please try again.";
      default:
        return err.rawMessage || "Something went wrong. Please try again.";
    }
  }
  if (err instanceof Error) {
    return err.message;
  }
  return "Something went wrong. Please try again.";
}
