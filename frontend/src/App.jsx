import React, { useState, useEffect } from "react";
import "bootstrap/dist/css/bootstrap.min.css";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Layout from "./pages/Layout";
import Home from "./pages/Home";
import NoPage from "./pages/NoPage";
import UserContext from "./context/UserContext";
import FingerprintJS from "@fingerprintjs/fingerprintjs";

const App = () => {
  const [visitorId, setVisitorId] = useState(null);

  useEffect(() => {
    FingerprintJS.load().then((fp) => {
      fp.get().then((result) => {
        setVisitorId(result.visitorId); 
        console.log(result.visitorId);  
      });
    });
  }, []); 

  if (!visitorId) {
    return <div>Loading...</div>; // Show a loading state while the visitorId is not ready
  }

  // Render the app once visitorId is available
  return (
    <UserContext.Provider value={visitorId}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Home />} />
            <Route path="*" element={<NoPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </UserContext.Provider>
  );
};

export default App;
