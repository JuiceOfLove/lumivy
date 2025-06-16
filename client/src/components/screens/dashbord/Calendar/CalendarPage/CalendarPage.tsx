import React, { useEffect, useState } from "react";
import { observer } from "mobx-react-lite";
import { useParams } from "react-router";
import EventService from "../../../../../services/EventService";
import { IEvent } from "../../../../../types/EventTypes";
import styles from "./CalendarPage.module.css";

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 0, 0, 0);
}
function endOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59);
}
function isEventOnDate(ev: IEvent, date: Date) {
  const evStart = new Date(ev.start_time);
  const evEnd = new Date(ev.end_time);
  return evStart <= endOfDay(date) && evEnd >= startOfDay(date);
}

const monthNames = [
  "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
  "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"
];

const CalendarPage: React.FC = observer(() => {
  const { id } = useParams();
  const calendarId = id ? parseInt(id, 10) : 1;

  const [currentMonth, setCurrentMonth] = useState<number>(new Date().getMonth());
  const [currentYear, setCurrentYear] = useState<number>(new Date().getFullYear());
  const [events, setEvents] = useState<IEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const [popupOpen, setPopupOpen] = useState<boolean>(false);
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);
  const [dayEvents, setDayEvents] = useState<IEvent[]>([]);
  const [showCreateForm, setShowCreateForm] = useState<boolean>(false);

  const [dayAllDay, setDayAllDay] = useState<boolean>(true);
  const [dayTitle, setDayTitle] = useState<string>("");
  const [dayDescription, setDayDescription] = useState<string>("");
  const [dayStartTime, setDayStartTime] = useState<string>("");
  const [dayEndTime, setDayEndTime] = useState<string>("");
  const [dayPrivate, setDayPrivate] = useState<boolean>(false);

  const [intervalModalOpen, setIntervalModalOpen] = useState<boolean>(false);
  const [intervalTitle, setIntervalTitle] = useState<string>("");
  const [intervalDesc, setIntervalDesc] = useState<string>("");
  const [intervalStart, setIntervalStart] = useState<string>("");
  const [intervalEnd, setIntervalEnd] = useState<string>("");
  const [intervalColor, setIntervalColor] = useState<string>("#ff5252");
  const [intervalPrivate, setIntervalPrivate] = useState<boolean>(false);

  const yearOptions = Array.from(
    { length: 11 },
    (_, i) => new Date().getFullYear() - 5 + i
  );

  useEffect(() => {
    loadEvents();
  }, [currentMonth, currentYear, calendarId]);

  async function loadEvents() {
    setLoading(true);
    setError("");
    try {
      const data = await EventService.getEventsForCalendar(
        calendarId,
        currentMonth + 1,
        currentYear
      );
      setEvents(data);
    } catch (e: any) {
      setError("Ошибка при загрузке");
    } finally {
      setLoading(false);
    }
  }

  function prevMonth() {
    let m = currentMonth - 1, y = currentYear;
    if (m < 0) { m = 11; y -= 1; }
    setCurrentMonth(m);
    setCurrentYear(y);
    closeDayPopup();
  }
  function nextMonth() {
    let m = currentMonth + 1, y = currentYear;
    if (m > 11) { m = 0; y += 1; }
    setCurrentMonth(m);
    setCurrentYear(y);
    closeDayPopup();
  }
  function onMonthChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setCurrentMonth(parseInt(e.target.value, 10));
    closeDayPopup();
  }
  function onYearChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setCurrentYear(parseInt(e.target.value, 10));
    closeDayPopup();
  }

  function handleDayClick(dayNum: number) {
    const date = new Date(currentYear, currentMonth, dayNum);
    const daily = events.filter(ev => isEventOnDate(ev, date));
    setDayEvents(daily);
    setSelectedDate(date);
    setPopupOpen(true);
    setShowCreateForm(false);

    const yyyy = date.getFullYear();
    const mm = String(date.getMonth() + 1).padStart(2, "0");
    const dd = String(date.getDate()).padStart(2, "0");
    setDayAllDay(true);
    setDayTitle("");
    setDayDescription("");
    setDayStartTime(`${yyyy}-${mm}-${dd}T00:00`);
    setDayEndTime(`${yyyy}-${mm}-${dd}T23:59`);
    setDayPrivate(false);
  }
  function closeDayPopup() {
    setPopupOpen(false);
    setSelectedDate(null);
    setDayEvents([]);
    setShowCreateForm(false);
  }

  async function handleIntervalSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!intervalStart || !intervalEnd) return;
    await EventService.createEvent({
      calendar_id: calendarId,
      title: intervalTitle,
      description: intervalDesc,
      start_time: new Date(intervalStart).toISOString(),
      end_time: new Date(intervalEnd).toISOString(),
      color: intervalColor,
      private: intervalPrivate,
    });
    setIntervalModalOpen(false);
    loadEvents();
  }

  async function handleAddDayEvent(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedDate) return;
    let s = new Date(dayStartTime), e2 = new Date(dayEndTime);
    if (dayAllDay) {
      s = startOfDay(selectedDate);
      e2 = endOfDay(selectedDate);
    }
    await EventService.createEvent({
      calendar_id: calendarId,
      title: dayTitle,
      description: dayDescription,
      start_time: s.toISOString(),
      end_time: e2.toISOString(),
      color: "",
      private: dayPrivate,
    });
    loadEvents();
    const updated = events.filter(ev => isEventOnDate(ev, selectedDate));
    setDayEvents(updated);
    setShowCreateForm(false);
  }

  async function handleCompleteEvent(evId: number) {
    await EventService.completeEvent(evId);
    loadEvents();
    if (selectedDate) {
      setDayEvents(prev =>
        prev.map(ev => (ev.id === evId ? { ...ev, is_completed: true } : ev))
      );
    }
  }

  function getDayIndicators(dayNum: number) {
    const date = new Date(currentYear, currentMonth, dayNum);
    const dayEvs = events.filter(ev => isEventOnDate(ev, date));
    const intervalColors: string[] = [];
    let hasNormal = false;
    dayEvs.forEach(ev => {
      if (ev.color) {
        if (!intervalColors.includes(ev.color)) {
          intervalColors.push(ev.color);
        }
      } else {
        hasNormal = true;
      }
    });
    return { intervalColors, hasNormal };
  }

  const daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate();
  const firstDayIdx = new Date(currentYear, currentMonth, 1).getDay();
  const calendarCells: (number | null)[] = [];
  for (let i = 0; i < firstDayIdx; i++) calendarCells.push(null);
  for (let d = 1; d <= daysInMonth; d++) calendarCells.push(d);

  if (loading) return <div className={styles.loading}>Загрузка...</div>;

  return (
    <div className={styles.calendarContainer}>
      <div className={styles.header}>
        <div className={styles.navControls}>
          <button onClick={prevMonth} className={styles.navButton}>←</button>
          <select value={currentMonth} onChange={onMonthChange} className={styles.selectMonth}>
            {monthNames.map((m, i) => <option key={i} value={i}>{m}</option>)}
          </select>
          <select value={currentYear} onChange={onYearChange} className={styles.selectYear}>
            {yearOptions.map(y => <option key={y} value={y}>{y}</option>)}
          </select>
          <button onClick={nextMonth} className={styles.navButton}>→</button>
        </div>
        <button
          onClick={() => { setIntervalModalOpen(true); setIntervalPrivate(false); }}
          className={styles.createIntervalBtn}
        >
          + Интервальное
        </button>
      </div>

      {error && <div className={styles.error}>{error}</div>}
      <div className={styles.weekdays}>
        {["Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"].map(wd => (
          <div key={wd} className={styles.weekdayCell}>{wd}</div>
        ))}
      </div>

      <div className={styles.grid}>
        {calendarCells.map((val, idx) => {
          if (val === null) {
            return <div key={idx} className={styles.emptyCell} />;
          }
          const isToday = (
            val === new Date().getDate() &&
            currentMonth === new Date().getMonth() &&
            currentYear === new Date().getFullYear()
          );
          const { intervalColors, hasNormal } = getDayIndicators(val);
          return (
            <div
              key={idx}
              className={`${styles.dayCell} ${isToday ? styles.today : ""}`}
              onClick={() => handleDayClick(val)}
            >
              <span className={styles.dayNumber}>{val}</span>
              {intervalColors.map((clr, i) => (
                <span
                  key={i}
                  className={styles.intervalDot}
                  style={{ backgroundColor: clr }}
                />
              ))}
              {hasNormal && <span className={styles.normalDot} />}
            </div>
          );
        })}
      </div>

      {intervalModalOpen && (
        <div className={styles.modalBackdrop}>
          <div className={styles.modal}>
            <h3 className={styles.modalTitle}>Новое интервальное событие</h3>
            <form onSubmit={handleIntervalSubmit} className={styles.modalForm}>
              <label>
                <span>Название:</span>
                <input
                  className={styles.modalInput}
                  value={intervalTitle}
                  onChange={e => setIntervalTitle(e.target.value)}
                  required
                />
              </label>
              <label>
                <span>Описание:</span>
                <textarea
                  className={styles.modalTextarea}
                  value={intervalDesc}
                  onChange={e => setIntervalDesc(e.target.value)}
                />
              </label>
              <label>
                <span>Начало:</span>
                <input
                  type="datetime-local"
                  className={styles.modalInput}
                  value={intervalStart}
                  onChange={e => setIntervalStart(e.target.value)}
                  required
                />
              </label>
              <label>
                <span>Окончание:</span>
                <input
                  type="datetime-local"
                  className={styles.modalInput}
                  value={intervalEnd}
                  onChange={e => setIntervalEnd(e.target.value)}
                  required
                />
              </label>
              <label>
                <span>Цвет:</span>
                <input
                  type="color"
                  className={styles.modalColorInput}
                  value={intervalColor}
                  onChange={e => setIntervalColor(e.target.value)}
                />
              </label>
              <label className={styles.allDayLabel}>
                <input
                  type="checkbox"
                  checked={intervalPrivate}
                  onChange={e => setIntervalPrivate(e.target.checked)}
                /> Приватное
              </label>
              <div className={styles.modalButtons}>
                <button type="submit" className={styles.saveBtn}>Создать</button>
                <button
                  type="button"
                  onClick={() => setIntervalModalOpen(false)}
                  className={styles.cancelBtn}
                >
                  Отмена
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {popupOpen && selectedDate && (
        <div className={styles.dayPopupBackdrop}>
          <div className={styles.dayPopupWindow}>
            <button className={styles.closeBtn} onClick={closeDayPopup}>✕</button>
            <h3 className={styles.popupTitle}>
              События:{" "}
              {`${selectedDate.getDate()}.${selectedDate.getMonth() + 1}.${selectedDate.getFullYear()}`}
            </h3>
            <div className={styles.eventsScrollArea}>
              {dayEvents.length === 0 ? (
                <p className={styles.noEvents}>Событий нет</p>
              ) : (
                <ul className={styles.eventList}>
                  {dayEvents.map(ev => (
                    <li key={ev.id} className={styles.eventItem}>
                      <div className={styles.eventHeader}>
                        <strong>{ev.title}</strong>
                        {ev.color && (
                          <span
                            className={styles.eventColorDot}
                            style={{ backgroundColor: ev.color }}
                          />
                        )}
                      </div>
                      <div className={styles.eventTime}>
                        {`${new Date(ev.start_time).toLocaleString()} – ${new Date(
                          ev.end_time
                        ).toLocaleString()}`}
                      </div>
                      <div className={styles.eventMeta}>
                        Автор: {ev.created_by}{" "}
                        {ev.is_completed ? (
                          <span className={styles.completedLabel}>(Выполнено)</span>
                        ) : (
                          <button
                            onClick={() => handleCompleteEvent(ev.id)}
                            className={styles.completeBtn}
                          >
                            Завершить
                          </button>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <hr className={styles.separator} />
            {!showCreateForm ? (
              <div className={styles.addEventWrapper}>
                <button
                  onClick={() => { setShowCreateForm(true); setDayPrivate(false); }}
                  className={styles.addEventBtn}
                >
                  + Добавить событие
                </button>
              </div>
            ) : (
              <>
                <h4 className={styles.subTitle}>Добавить событие на этот день</h4>
                <form onSubmit={handleAddDayEvent} className={styles.modalForm}>
                  <label>
                    <span>Название:</span>
                    <input
                      className={styles.modalInput}
                      value={dayTitle}
                      onChange={e => setDayTitle(e.target.value)}
                      required
                    />
                  </label>
                  <label>
                    <span>Описание:</span>
                    <textarea
                      className={styles.modalTextarea}
                      value={dayDescription}
                      onChange={e => setDayDescription(e.target.value)}
                    />
                  </label>
                  <label className={styles.allDayLabel}>
                    <input
                      type="checkbox"
                      checked={dayAllDay}
                      onChange={e => setDayAllDay(e.target.checked)}
                    /> Весь день
                  </label>
                  {!dayAllDay && (
                    <>
                      <label>
                        <span>Начало:</span>
                        <input
                          type="datetime-local"
                          className={styles.modalInput}
                          value={dayStartTime}
                          onChange={e => setDayStartTime(e.target.value)}
                        />
                      </label>
                      <label>
                        <span>Окончание:</span>
                        <input
                          type="datetime-local"
                          className={styles.modalInput}
                          value={dayEndTime}
                          onChange={e => setDayEndTime(e.target.value)}
                        />
                      </label>
                    </>
                  )}
                  <label className={styles.allDayLabel}>
                    <input
                      type="checkbox"
                      checked={dayPrivate}
                      onChange={e => setDayPrivate(e.target.checked)}
                    /> Приватное
                  </label>
                  <div className={styles.modalButtons}>
                    <button type="submit" className={styles.saveBtn}>Сохранить</button>
                    <button
                      type="button"
                      onClick={() => setShowCreateForm(false)}
                      className={styles.cancelBtn}
                    >
                      Отмена
                    </button>
                  </div>
                </form>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
});

export default CalendarPage;
