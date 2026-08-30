import { message } from "./message";

export const handler = async (): Promise<{ statusCode: number; body: string }> => ({
  statusCode: 200,
  body: message,
});
