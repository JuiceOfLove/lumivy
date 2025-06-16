import { useEffect, useState } from "react";
import AdminService from "../../../../services/AdminService";
import { IPayment } from "../../../../types/payment";
import { IUser } from "../../../../types/auth";
import styles from "./AdminPanel.module.css";

const useDebounce = (value: string, ms = 300) => {
  const [deb, setDeb] = useState(value);
  useEffect(() => {
    const h = setTimeout(() => setDeb(value), ms);
    return () => clearTimeout(h);
  }, [value, ms]);
  return deb;
};

const AdminPanel = () => {
  const [payments, setPayments] = useState<IPayment[]>([]);
  const [loadingPay, setLoadingPay] = useState(true);

  const [operators, setOperators] = useState<IUser[]>([]);
  const [loadingOps, setLoadingOps] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [query, setQuery] = useState("");
  const debQuery = useDebounce(query, 300);
  const [suggests, setSuggests] = useState<IUser[]>([]);
  const [searching, setSearching] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const pay = await AdminService.getPaymentHistory();
        setPayments(pay);
      } finally { setLoadingPay(false); }
    })();
  }, []);

  useEffect(() => {
    loadOperators();
  }, []);

  async function loadOperators() {
    setLoadingOps(true);
    try {
      const ops = await AdminService.listOperators();
      setOperators(ops);
    } finally { setLoadingOps(false); }
  }

  useEffect(() => {
    if (!debQuery.trim()) { setSuggests([]); return; }
    (async () => {
      setSearching(true);
      try {
        const res = await AdminService.searchUsers(debQuery.trim());
        setSuggests(res);
      } finally { setSearching(false); }
    })();
  }, [debQuery]);

  const makeOp = async (id: number) => {
    await AdminService.makeOperator(id);
    setModalOpen(false);
    setQuery("");
    await loadOperators();
  };

  return (
    <div className={styles.adminContainer}>
      <h1>Админ-панель</h1>

      <section className={styles.opsSection}>
        <div className={styles.opsHeader}>
          <h2>Операторы</h2>
          <button onClick={() => setModalOpen(true)} className={styles.addBtn}>
            + Добавить оператора
          </button>
        </div>

        {loadingOps ? <p>Загрузка списка…</p> :
          operators.length === 0 ?
            <p>Пока операторов нет</p> :
            <ul className={styles.opsList}>
              {operators.map(op => (
                <li key={op.id}>
                  <strong>{op.name}</strong> — {op.email}
                  <span className={styles.opDate}>
                    {new Date(op.created_at!).toLocaleDateString()}
                  </span>
                </li>
              ))}
            </ul>}
      </section>

      <section>
        <h2>История транзакций</h2>
        {loadingPay ? (
          <p>Загрузка…</p>
        ) : payments.length === 0 ? (
          <p>Транзакций не найдено</p>
        ) : (
          <table className={styles.table}>
            <thead>
              <tr>
                <th>ID</th><th>Payment ID</th><th>Семья</th><th>Пользователь</th>
                <th>Сумма</th><th>Статус</th><th>Дата</th>
              </tr>
            </thead>
            <tbody>
              {payments.map(p => (
                <tr key={p.id}>
                  <td>{p.id}</td><td>{p.payment_id}</td><td>{p.family_id}</td>
                  <td>{p.user_id}</td><td>{p.amount}</td><td>{p.status}</td>
                  <td>{new Date(p.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
      {modalOpen && (
        <div className={styles.modalBackdrop} onClick={() => setModalOpen(false)}>
          <div className={styles.modal} onClick={e => e.stopPropagation()}>
            <h3>Назначить оператора</h3>

            <input
              className={styles.modalInput}
              placeholder="Введите e-mail…"
              value={query}
              onChange={e => setQuery(e.target.value)}
            />

            {searching && <p>Поиск…</p>}
            {!searching && suggests.length === 0 && debQuery && (
              <p>Ничего не найдено</p>
            )}

            <ul className={styles.suggestList}>
              {suggests.map(u => (
                <li key={u.id} onClick={() => makeOp(u.id)}>
                  {u.name || "—"} <span>{u.email}</span>
                </li>
              ))}
            </ul>

            <button onClick={() => setModalOpen(false)} className={styles.cancelBtn}>
              Закрыть
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminPanel;
