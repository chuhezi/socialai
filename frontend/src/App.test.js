import { render, screen } from "@testing-library/react";
import { BrowserRouter } from "react-router-dom";
import App from "./components/App";

jest.mock("axios", () => ({
  __esModule: true,
  default: jest.fn(),
}));

jest.mock("./components/PhotoGallery", () => () => null);
jest.mock("./components/Landing", () => () => <div>Social AI</div>);

test("renders the login page", () => {
  localStorage.clear();
  render(
    <BrowserRouter>
      <App />
    </BrowserRouter>,
  );
  expect(screen.getByRole("button", { name: /log in/i })).toBeInTheDocument();
});
