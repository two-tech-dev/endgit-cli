import axios from "axios";
import { getConfig } from "./config";

export function getApiClient() {
  const config = getConfig();
  const baseURL = config.apiUrl || "http://localhost:4000";
  
  const client = axios.create({
    baseURL,
    headers: {
      "Content-Type": "application/json"
    }
  });

  if (config.apiToken) {
    client.defaults.headers.common["Authorization"] = `Bearer ${config.apiToken}`;
  }

  return client;
}
