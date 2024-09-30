import { useEffect } from "react";
import DealsCarousel from "../components/DealsCarousel";
import CookieConsent from "../components/CookieConsent";

const Home = () => {
  useEffect(() => {}, []);

  return (
    <div>
      <h2>Welcome to our Deals</h2>
      <DealsCarousel />
      <CookieConsent />
    </div>
  );
};

export default Home;
