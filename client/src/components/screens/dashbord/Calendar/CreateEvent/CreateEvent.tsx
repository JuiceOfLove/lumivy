import React, { useState } from "react";
import { useNavigate } from "react-router";
import EventService from "../../../../../services/EventService";

const CreateEvent: React.FC = () => {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await EventService.createEvent({
      title,
      description,
      start_time: new Date(startTime).toISOString(),
      end_time: new Date(endTime).toISOString(),
      private: isPrivate,
    });
    navigate("/dashboard");
  };

  return (
    <div>
      <h2>Создать событие</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label>Название:</label>
          <input value={title} onChange={e => setTitle(e.target.value)} required />
        </div>
        <div>
          <label>Описание:</label>
          <textarea value={description} onChange={e => setDescription(e.target.value)} />
        </div>
        <div>
          <label>Начало:</label>
          <input type="datetime-local" onChange={e => setStartTime(e.target.value)} required />
        </div>
        <div>
          <label>Окончание:</label>
          <input type="datetime-local" onChange={e => setEndTime(e.target.value)} required />
        </div>
        <div>
          <label>
            <input
              type="checkbox"
              checked={isPrivate}
              onChange={e => setIsPrivate(e.target.checked)}
            />{" "}
            Приватное
          </label>
        </div>
        <button type="submit">Создать</button>
      </form>
    </div>
  );
};

export default CreateEvent;