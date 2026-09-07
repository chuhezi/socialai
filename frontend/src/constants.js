export const TOKEN_KEY = "token";

export const BASE_URL =
  process.env.REACT_APP_API_BASE_URL ||
  "http://127.0.0.1:8080";

export const SEARCH_KEY = {
  all: 0,
  keyword: 1,
  user: 2,
  mine: 3,
};
