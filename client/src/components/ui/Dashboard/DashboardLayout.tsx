import React, { useContext, useEffect, useState } from "react";
import { Outlet, Link, useNavigate } from "react-router";
import { observer } from "mobx-react-lite";
import { Context } from "../../../main";
import SubscriptionService from "../../../services/SubscriptionService";
import CalendarService from "../../../services/CalendarService";
import styles from "./DashboardLayout.module.css";

const DashboardLayout: React.FC = () => {
  const { store } = useContext(Context);
  const navigate = useNavigate();

  const [isPremium, setIsPremium] = useState(false);
  const [calendars, setCalendars] = useState<any[]>([]);
  const [calOpen, setCalOpen] = useState(false);

  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [newCalendarTitle, setNewCalendarTitle] = useState("");

  useEffect(() => {
    if (!store.authChecked) return;
    if (!store.isAuth) navigate("/auth/login");
    else if (store.user && store.user.family_id === 0)
      navigate("/dashboard/family/create");
    else {
      checkSubscriptionStatus();
      loadCalendars();
    }
  }, [store.isAuth, store.authChecked, store.user, navigate]);

  async function checkSubscriptionStatus() {
    try {
      const data = await SubscriptionService.checkSubscription();
      setIsPremium(data.isActive);
    } catch (err) {
      console.error("Ошибка при проверке подписки:", err);
    }
  }

  async function loadCalendars() {
    try {
      const cals = await CalendarService.getAllCalendars();
      setCalendars(cals);
    } catch (err) {
      console.error("Ошибка загрузки календарей:", err);
    }
  }

  function toggleCalOpen() {
    setCalOpen(!calOpen);
  }

  async function handleLogout() {
    await store.logout();
    navigate("/auth/login");
  }

  const handleBuySubscription = async () => {
    try {
      const data = await SubscriptionService.buySubscription();
      if (data.payment_url) {
        window.location.href = data.payment_url;
      }
    } catch (err) {
      console.error("Ошибка при покупке подписки:", err);
      alert("Не удалось инициировать оплату. Проверьте консоль.");
    }
  };

  const openCreateModal = () => {
    setNewCalendarTitle("");
    setCreateModalOpen(true);
  };
  const closeCreateModal = () => {
    setCreateModalOpen(false);
  };
  const submitCreateCalendar = async () => {
    if (!newCalendarTitle.trim()) return;
    try {
      await CalendarService.createExtraCalendar(newCalendarTitle.trim());
      await loadCalendars();
      closeCreateModal();
    } catch (err) {
      console.error("Ошибка создания календаря:", err);
    }
  };

  return (
    <div className={styles.dashboardContainer}>
      <header className={styles.dashboardHeader}>
        <div className={styles.logo}>Lumivy</div>
      </header>

      <div className={styles.body}>
        <aside className={styles.sidebar}>
          {isPremium ? (
            <div className={styles.premiumBanner}>Премиум пользователь</div>
          ) : (
            <div className={styles.normalBanner}>
              Обычный пользователь
              <button
                className={styles.buyBtn}
                onClick={handleBuySubscription}
              >
                Купить подписку
              </button>
            </div>
          )}

          <nav>
            <ul className={styles.navList}>
              <li>
                <Link to="/dashboard" className={styles.navLink}>
                  Главная
                </Link>
              </li>
              <li>
                <div
                  onClick={toggleCalOpen}
                  className={styles.navLink}
                  style={{ cursor: "pointer" }}
                >
                  Календари {calOpen ? "▲" : "▼"}
                </div>
                {calOpen && (
                  <ul className={styles.subNavList}>
                    {calendars.map((cal) => (
                      <li key={cal.id}>
                        <Link
                          to={`/dashboard/calendar/${cal.id}`}
                          className={styles.navLink}
                        >
                          {cal.title || "Календарь #" + cal.id}
                        </Link>
                      </li>
                    ))}
                    {isPremium && (
                      <li>
                        <button
                          onClick={openCreateModal}
                          className={styles.navLinkBtn}
                        >
                          + Создать календарь
                        </button>
                      </li>
                    )}
                  </ul>
                )}
              </li>
              <li>
                <Link to="/dashboard/chat" className={styles.navLink}>
                  Чат
                </Link>
              </li>
              <li>
                <Link to="/dashboard/support" className={styles.navLink}>
                  Поддержка
                </Link>
              </li>
              <li>
                <Link to="/dashboard/profile" className={styles.navLink}>
                  Профиль
                </Link>
              </li>
              <li>
                <Link to="/dashboard/family" className={styles.navLink}>
                  Семья
                </Link>
              </li>
              {store.user?.role === "operator" && (
                <li>
                  <Link
                    to="/dashboard/support/operator/new"
                    className={styles.navLink}
                  >
                    Тикеты (оператор)
                  </Link>
                </li>
              )}
              {store.user?.role === "admin" && (
                <li>
                  <Link to="/dashboard/admin" className={styles.navLink}>
                    Админка
                  </Link>
                </li>
              )}
              <li>
                <span
                  onClick={handleLogout}
                  className={styles.navLink}
                  style={{ cursor: "pointer" }}
                >
                  Выйти
                </span>
              </li>
            </ul>
          </nav>
        </aside>

        <main className={styles.mainContent}>
          <Outlet />
        </main>
      </div>

      {/* Mobile bottom navigation */}
      <footer className={styles.mobileNav}>
        <Link to="/dashboard" className={styles.mobileNavLink}>
          🏠
        </Link>
        <Link to="/dashboard/calendar" className={styles.mobileNavLink}>
          📅
        </Link>
        <Link to="/dashboard/chat" className={styles.mobileNavLink}>
          💬
        </Link>
        <Link to="/dashboard/support" className={styles.mobileNavLink}>
          🆘
        </Link>
        <Link to="/dashboard/profile" className={styles.mobileNavLink}>
          👤
        </Link>
      </footer>
      {createModalOpen && (
        <div className={styles.modalBackdrop}>
          <div className={styles.modal}>
            <h3 className={styles.modalTitle}>Новый календарь</h3>
            <input
              type="text"
              className={styles.modalInput}
              placeholder="Название календаря"
              value={newCalendarTitle}
              onChange={(e) => setNewCalendarTitle(e.target.value)}
              autoFocus
            />
            <div className={styles.modalButtons}>
              <button
                onClick={submitCreateCalendar}
                className={styles.saveBtn}
              >
                Создать
              </button>
              <button onClick={closeCreateModal} className={styles.cancelBtn}>
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default observer(DashboardLayout);