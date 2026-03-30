import { createRouter, createWebHistory } from 'vue-router';

const routes = [
  { path: '/', redirect: '/usage' },
  { path: '/usage', component: { template: '<div></div>' } },
  { path: '/tags', component: { template: '<div></div>' } },
  { path: '/providers', component: { template: '<div></div>' } },
  { path: '/keys', component: { template: '<div></div>' } },
  { path: '/stats', component: { template: '<div></div>' } },
  { path: '/logs', component: { template: '<div></div>' } },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
