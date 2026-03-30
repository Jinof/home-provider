import { createApp } from 'vue';
import { i18n } from './i18n';
import { router } from './router';
import App from './App.vue';

window.onerror = (msg, src, line, col, err) => {
  console.error('[GLOBAL ERROR]', msg, src, line, col, err);
  document.body.innerHTML =
    '<div style="padding:20px;font-family:monospace;color:red"><h2>JS Error</h2><pre>' +
    msg +
    '\n' +
    (err?.stack || '') +
    '</pre></div>';
  return true;
};

const app = createApp(App);
app.config.errorHandler = (err, instance, info) => {
  console.error('[VUE ERROR]', err, info);
  document.body.innerHTML =
    '<div style="padding:20px;font-family:monospace;color:red"><h2>Vue Error</h2><pre>' +
    err +
    '\n' +
    info +
    '</pre></div>';
};
app.use(i18n).use(router).mount('#app');
