import { Link } from "react-router";
import styles from "./PaymentSuccess.module.css";

const PaymentSuccess: React.FC = () => (
  <div className={styles.successContainer}>
    <div className={styles.successCard}>
      <div className={styles.content}>
        <h2>Оплата прошла успешно!</h2>
        <p>Спасибо за покупку подписки.</p>

        <Link to="/dashboard" className={styles.linkBtn}>
          Вернуться в Lumivy
        </Link>
      </div>
    </div>
  </div>
);

export default PaymentSuccess;