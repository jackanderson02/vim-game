import "bootstrap/dist/css/bootstrap.min.css";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Layout from "./pages/Layout";
import Home from "./pages/Home";
import NoPage from "./pages/NoPage";
import UserContext from "./context/IDContext";
import FingerprintJS from '@fingerprintjs/fingerprintjs';


// import Products from "./pages/Products";
// import Login from "./pages/Login";
// import SignUp from "./pages/SignUp.jsx";

const App = () => {
  let visitorId
  FingerprintJS.load().then(fp => {
    fp.get().then(result => {
        visitorId = result.visitorId;
        console.log(visitorId); // Logs a unique fingerprint ID for the device
    });
  });
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
