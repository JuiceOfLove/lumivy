import { BrowserRouter, Route, Routes, useLocation } from "react-router";
import { useContext, useEffect } from "react";
import { observer } from "mobx-react-lite";
import { Context } from "../../main";

import NotFound from "../screens/404/NotFound";
import Layout from "../ui/Layout/Layout";
import Home from "../screens/Home/Home";
import Login from "../screens/Auth/Login/Login";
import Registration from "../screens/Auth/Registration/Registration";
import Activate from "../screens/Auth/Activate/Activate"; // Ваш компонент главной страницы дашборда
import AcceptInvitation from "../screens/dashbord/Family/Accept/AcceptInvitation";
import CreateFamily from "../screens/dashbord/Family/Create/CreateFamily";
import DashboardLayout from "../ui/Dashboard/DashboardLayout";
import DashboardHome from "../screens/dashbord/Home/DashboardHome";
import CalendarPage from "../screens/dashbord/Calendar/CalendarPage/CalendarPage";
import CreateEvent from "../screens/dashbord/Calendar/CreateEvent/CreateEvent";
import CalendarView from "../screens/dashbord/Calendar/CalendarView/CalendarView";
import PaymentSuccess from "../screens/dashbord/PaymentSuccess/PaymentSuccess";
import Profile from "../screens/dashbord/Profile/Profile";
import FamilyDetails from "../screens/dashbord/Family/Details/FamilyDetails";
import AdminPanel from "../screens/dashbord/Admin/AdminPage";
import ChatPage from "../screens/dashbord/chat/ChatPage/ChatPage";
import MyTicketsList from "../screens/dashbord/Support/MyTicketsList/MyTicketsList";
import CreateTicket from "../screens/dashbord/Support/CreateTicket/CreateTicket";
import TicketChat from "../screens/dashbord/Support/TicketChat/TicketChat";
import OperatorTickets from "../screens/dashbord/Support/OperatorTickets/OperatorTickets";


const AppRoutes = () => {
  const { store } = useContext(Context);
  const location = useLocation();

  useEffect(() => {
    if (
      !location.pathname.includes("/auth/activate") &&
      !location.pathname.includes("/family/invite")
    ) {
      if (localStorage.getItem("token")) {
        store.checkAuth();
      }
    }
  }, [store, location.pathname]);

  if (store.isLoading) {
    return <div>Загрузка...</div>;
  }

  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Home />} />
        <Route path="auth/login" element={<Login />} />
        <Route path="auth/register" element={<Registration />} />
        <Route path="auth/activate/:link" element={<Activate />} />
        <Route path="*" element={<NotFound />} />
      </Route>

      <Route path="dashboard" element={<DashboardLayout />}>
        <Route index element={<DashboardHome />} />
        <Route path="family/create" element={<CreateFamily />} />
        <Route path="family/invite/:token" element={<AcceptInvitation />} />
        <Route path="family" element={<FamilyDetails />} />

        <Route path="calendar/create" element={<CreateEvent />} />
        <Route path="calendar/view" element={<CalendarView />} />
        <Route path="calendar/:id" element={<CalendarPage />} />
        <Route path="payment-success" element={<PaymentSuccess />} />

        <Route path="chat" element={<ChatPage />} />

        <Route path="support" element={<MyTicketsList />} />
        <Route path="support/new" element={<CreateTicket />} />
        <Route path="support/ticket/:id" element={<TicketChat />} />
        <Route path="support/operator/:status" element={<OperatorTickets />} />
        <Route path="support/operator/ticket/:id" element={<TicketChat />} />

        <Route path="profile" element={<Profile />} />

        <Route path="admin" element={<AdminPanel />} />
      </Route>
    </Routes>
  );
};

export default observer(() => (
  <BrowserRouter>
    <AppRoutes />
  </BrowserRouter>
));