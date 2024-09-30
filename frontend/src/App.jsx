import "bootstrap/dist/css/bootstrap.min.css";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Layout from "./pages/Layout";
import Home from "./pages/Home";
import About from "./pages/About";
import Contact from "./pages/Contact";
import NoPage from "./pages/NoPage";
import Basket from "./pages/Basket";
import Checkout from "./pages/Checkout";
import ProductList from "./components/ProductList";
import OrderConfirmation from "./pages/OrderConfirmation";
import { useState } from "react";
import { BasketContext } from "./context/BasketContext";
import { LoginContext } from "./context/LoginContext.jsx";

import Cookies from "js-cookie";
import Products from "./pages/Products";
import Login from "./pages/Login";
import SignUp from "./pages/SignUp.jsx";

const App = () => {
  let initBasketState;

  const basketCookie = Cookies.get("basket");

  if (basketCookie) {
    initBasketState = JSON.parse(basketCookie);
  } else {
    initBasketState = {};
  }

  const [basket, setBasket] = useState(initBasketState);
  const [loggedIn, setLoggedIn] = useState(false);


  let initSelectedDeals;
  const dealsCookie = Cookies.get("deals");

  if (dealsCookie) {
    initSelectedDeals = JSON.parse(dealsCookie);
  } else {
    initSelectedDeals = {};
  }

  //eslint-disable-next-line no-unused-vars
  const [selectedDeals, setSelectedDeals] = useState(initSelectedDeals);

  return (
    <LoginContext.Provider value={[loggedIn, setLoggedIn]}>
      <BasketContext.Provider value={{ basket, setBasket, selectedDeals, setSelectedDeals }}>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<Home />} />
              <Route path="about" element={<About />} />
              <Route path="contact" element={<Contact />} />
              <Route
                path="basket"
                element={<Basket displayCheckout={true} />}
              />
              <Route path="checkout" element={<Checkout />} />
              <Route path="Products" element={<Products />} />
              <Route path="login" element={<Login />} />
              <Route path="signup" element={<SignUp />} />
              <Route path="orderconfirmation" element={<OrderConfirmation />} />
              <Route path="*" element={<NoPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </BasketContext.Provider>
    </LoginContext.Provider>
  );
};

export default App;
