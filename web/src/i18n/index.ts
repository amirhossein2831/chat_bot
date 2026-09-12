import i18n from "i18next";
import { initReactI18next } from "react-i18next";

import en from "./en.json";
import fa from "./fa.json";

void i18n.use(initReactI18next).init({
  resources: { fa: { translation: fa }, en: { translation: en } },
  lng: "fa",
  fallbackLng: "fa",
  interpolation: { escapeValue: false },
});

export default i18n;
