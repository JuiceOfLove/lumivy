import { useEffect, useState } from "react";
import EventService from "../../../../services/EventService";
import { IEvent } from "../../../../types/EventTypes";
import styles from "./DashboardHome.module.css";

function isEventOnDate(ev: IEvent, date: Date) {
  const start = new Date(ev.start_time).getTime();
  const end = new Date(ev.end_time).getTime();
  const dayStart = new Date(date.getFullYear(), date.getMonth(), date.getDate(), 0, 0, 0).getTime();
  const dayEnd = new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59).getTime();
  return start <= dayEnd && end >= dayStart;
}

const DashboardHome: React.FC = () => {
  const [allEvents, setAllEvents] = useState<IEvent[]>([]);
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [dayEvents, setDayEvents] = useState<IEvent[]>([]);
  const [upcomingEvents, setUpcomingEvents] = useState<IEvent[]>([]);

  useEffect(() => {
    (async () => {
      const events = await EventService.getAllEvents();
      setAllEvents(events);
    })();
  }, []);

  useEffect(() => {
    const de = allEvents
      .filter(ev => isEventOnDate(ev, selectedDate))
      .sort((a, b) => a.start_time.localeCompare(b.start_time));
    setDayEvents(de);

    const now = new Date().getTime();
    const future = allEvents
      .filter(ev => new Date(ev.start_time).getTime() > now)
      .sort((a, b) => a.start_time.localeCompare(b.start_time))
      .slice(0, 3);
    setUpcomingEvents(future);
  }, [selectedDate, allEvents]);

  async function handleComplete(id: number) {
    await EventService.completeEvent(id);
    setAllEvents(prev =>
      prev.map(ev => (ev.id === id ? { ...ev, is_completed: true } : ev))
    );
  }

  function prevDay() {
    setSelectedDate(d => new Date(d.getFullYear(), d.getMonth(), d.getDate() - 1));
  }
  function nextDay() {
    setSelectedDate(d => new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1));
  }

  const formattedDate = `${selectedDate.getDate()}.${selectedDate.getMonth() + 1}.${selectedDate.getFullYear()}`;

  return (
    <div className={styles.container}>
      <h1 className={styles.pageTitle}>Добро пожаловать в Lumivy</h1>

      <div className={styles.dateNav}>
        <button onClick={prevDay} className={styles.navBtn}>‹</button>
        <span className={styles.dateLabel}>{formattedDate}</span>
        <button onClick={nextDay} className={styles.navBtn}>›</button>
      </div>

      <section className={styles.section}>
        {dayEvents.length > 0 ? (
          <ul className={styles.eventList}>
            {dayEvents.map(ev => (
              <li key={ev.id} className={styles.eventItem}>
                <div className={styles.eventHeader}>
                  <span className={styles.eventTime}>
                    {new Date(ev.start_time).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                    {" – "}
                    {new Date(ev.end_time).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                  </span>
                  <strong className={styles.eventTitle}>{ev.title}</strong>
                </div>
                {ev.color && (
                  <span className={styles.colorDot} style={{ backgroundColor: ev.color }} />
                )}
                {!ev.is_completed ? (
                  <button
                    className={styles.completeBtn}
                    onClick={() => handleComplete(ev.id)}
                  >
                    Завершить
                  </button>
                ) : (
                  <span className={styles.completedLabel}>Выполнено</span>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p className={styles.noEvents}>На этот день событий нет.</p>
        )}
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Ближайшие события</h2>
        {upcomingEvents.length > 0 ? (
          <ul className={styles.upcomingList}>
            {upcomingEvents.map(ev => (
              <li key={ev.id} className={styles.upcomingItem}>
                <div className={styles.eventHeader}>
                  <span className={styles.eventTime}>
                    {new Date(ev.start_time).toLocaleString("ru-RU")}
                  </span>
                  <strong className={styles.eventTitle}>{ev.title}</strong>
                </div>
                {ev.color && (
                  <span className={styles.colorDot} style={{ backgroundColor: ev.color }} />
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p className={styles.noEvents}>Нет предстоящих событий.</p>
        )}
      </section>
    </div>
  );
};

export default DashboardHome;