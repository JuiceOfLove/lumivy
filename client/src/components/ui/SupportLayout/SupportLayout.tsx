import React, { useContext, useEffect } from "react";
import { Navigate, useNavigate } from "react-router";
import { Context } from "../../../main";

const SupportLayout: React.FC = () => {
  const { store } = useContext(Context);
  const navigate = useNavigate();

  useEffect(() => {
    if (!store.isAuth) {
      navigate("/auth/login");
    }
  }, [store.isAuth, navigate]);

  if (store.user?.role === "operator") {
    return <Navigate to="/dashboard/support/operator/new" replace />;
  } else {
    return <Navigate to="/dashboard/support" replace />;
  }
};

export default SupportLayout;