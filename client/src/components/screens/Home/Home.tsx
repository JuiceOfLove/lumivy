import { useEffect, useRef } from "react";
import { Link } from "react-router";
import gsap from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { ScrollSmoother } from "gsap/ScrollSmoother";

import styles from "./Home.module.css";

gsap.registerPlugin(ScrollTrigger, ScrollSmoother);


const Home = () => {
  const wrapper = useRef<HTMLDivElement>(null);
  const content = useRef<HTMLDivElement>(null);

  useEffect(() => {
    ScrollSmoother.create({
      wrapper: wrapper.current!,
      content: content.current!,
      smooth: 1.2,
      effects: true,
    });

    gsap.from(`.${styles.heroTitle}`, {
      y: -80,
      opacity: 0,
      duration: 1.2,
      ease: "power3.out",
    });
    gsap.from(`.${styles.heroSub}`, {
      y: 60,
      opacity: 0,
      duration: 1,
      delay: 0.25,
      ease: "power3.out",
    });
    gsap.from(`.${styles.cta}`, {
      scale: 0.8,
      opacity: 0,
      duration: 0.8,
      delay: 0.5,
      ease: "back.out(1.7)",
    });

    gsap.utils
      .toArray<HTMLImageElement>(`.${styles.parallaxImg}`)
      .forEach((img) => {
        const speed = Number(img.dataset.speed) || 1;
        gsap.to(img, {
          yPercent: speed * 100,
          ease: "none",
          scrollTrigger: {
            trigger: img,
            start: "top bottom",
            scrub: true,
          },
        });
      });
  }, []);

  return (
    <div ref={wrapper} className={styles.smoothWrapper}>
      <div ref={content} className={styles.smoothContent}>
        <section className={styles.heroSection}>
          <h1 className={`${styles.heroTitle} ${styles.gradientTxt}`}>
            Добро пожаловать в Lumivy
          </h1>

          <p className={styles.heroSub}>
            Планируйте мероприятия, ведите календари, общайтесь в семейном чате — всё в единой «лавандовой» экосистеме.
          </p>

          <Link to="/dashboard" className={`${styles.cta} ${styles.btnPurple}`}>
            Перейти в платформу
          </Link>
        </section>
        <section className={styles.featureSection}>
          <div className={styles.featureCard}>
            <h3>📅 Календари</h3>
            <p>
              Создавайте отдельные календари под проекты и семейные события.
            </p>
          </div>
          <div className={styles.featureCard}>
            <h3>💬 Чат</h3>
            <p>
              Отправляйте сообщения, делитесь картинками
            </p>
          </div>
          <div className={styles.featureCard}>
            <h3>🆘 Поддержка</h3>
            <p>
              Спросите оператора — получайте ответы и отслеживайте статус.
            </p>
          </div>
        </section>

        <section className={styles.bottomCTA}>
          <h2>Готовы попробовать?</h2>
          <Link
            to="/auth/register"
            className={`${styles.btnPurple} ${styles.ctaSecondary}`}
          >
            Создать аккаунт
          </Link>
        </section>
      </div>
    </div>
  );
};

export default Home;