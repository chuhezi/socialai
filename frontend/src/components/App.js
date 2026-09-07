import React, { useEffect, useState } from "react";
import { message } from "antd";
import ResponsiveAppBar from "./ResponsiveAppBar";
import Main from "./Main";
import { TOKEN_KEY } from "../constants";
import { request, session, errorText } from "../lib/api";

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(() => Boolean(session()));
  const [loggingOut, setLoggingOut] = useState(false);
  useEffect(() => {
    const sync = () => setIsLoggedIn(Boolean(session()));
    window.addEventListener("socialai:session-expired", sync);
    window.addEventListener("storage", sync);
    const timer = setInterval(sync, 10000);
    return () => { clearInterval(timer); window.removeEventListener("socialai:session-expired", sync); window.removeEventListener("storage", sync); };
  }, []);
  const logout = async () => {
    setLoggingOut(true);
    try {
      await request("/signout", { method: "POST" });
      localStorage.removeItem(TOKEN_KEY); setIsLoggedIn(false);
    } catch (error) { message.error(errorText(error, "Could not log out. Please retry.")); }
    finally { setLoggingOut(false); }
  };
  const loggedIn = token => {
    localStorage.setItem(TOKEN_KEY, token); setIsLoggedIn(Boolean(session()));
  };
  return <div className="App">
    <ResponsiveAppBar isLoggedIn={isLoggedIn} handleLogout={logout} loggingOut={loggingOut} />
    <Main isLoggedIn={isLoggedIn} handleLoggedIn={loggedIn} />
    <footer className="site-footer"><span>Social AI</span><span>A space for your imagination.</span></footer>
  </div>;
}
export default App;
