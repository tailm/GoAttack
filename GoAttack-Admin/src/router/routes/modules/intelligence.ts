import { DEFAULT_LAYOUT } from '../base'
import { AppRouteRecordRaw } from '../types'

const INTELLIGENCE: AppRouteRecordRaw = {
  path: '/intelligence',
  name: 'intelligence',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: 'menu.intelligence',
    icon: 'icon-search',
    requiresAuth: true,
    order: 3,
  },
  children: [
    {
      path: 'vulnerabilities',
      name: 'vulnerability-intelligence',
      component: () => import('@/views/intelligence/vulnerability/index.vue'),
      meta: {
        locale: 'menu.intelligence.vulnerabilities',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'detection-tasks',
      name: 'detection-tasks',
      component: () => import('@/views/intelligence/detection/index.vue'),
      meta: {
        locale: 'menu.intelligence.detectionTasks',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'alerts',
      name: 'alerts',
      component: () => import('@/views/intelligence/alert/index.vue'),
      meta: {
        locale: 'menu.intelligence.alerts',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'config',
      name: 'intelligence-config',
      component: () => import('@/views/intelligence/config/index.vue'),
      meta: {
        locale: 'menu.intelligence.config',
        requiresAuth: true,
        roles: ['admin'],
      },
    },
  ],
}

export default INTELLIGENCE