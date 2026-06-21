export const iconNames = [
  'analysis',
  'barChart',
  'check',
  'classroom',
  'dashboard',
  'delete',
  'distribution',
  'edit',
  'layer',
  'lineChart',
  'rank',
  'scores',
  'student',
  'sync',
  'teacher',
  'trend',
  'upload',
  'warning'
] as const

export type IconName = (typeof iconNames)[number]

export type IconDefinition = {
  paths: string[]
}

export const iconRegistry: Record<IconName, IconDefinition> = {
  analysis: {
    paths: [
      'M4 19V5',
      'M4 19H20',
      'M7 16L11 12L14 14L19 8',
      'M17 8H19V10'
    ]
  },
  barChart: {
    paths: [
      'M5 19V10',
      'M12 19V5',
      'M19 19V13',
      'M3 19H21'
    ]
  },
  check: {
    paths: [
      'M20 6L9 17L4 12',
      'M12 22C17.5 22 22 17.5 22 12C22 6.5 17.5 2 12 2C6.5 2 2 6.5 2 12C2 17.5 6.5 22 12 22Z'
    ]
  },
  classroom: {
    paths: [
      'M4 7L12 3L20 7L12 11L4 7Z',
      'M6 10V15C6 16.7 8.7 18 12 18C15.3 18 18 16.7 18 15V10',
      'M20 7V13'
    ]
  },
  dashboard: {
    paths: [
      'M4 13C4 8.6 7.6 5 12 5C16.4 5 20 8.6 20 13',
      'M12 13L16 9',
      'M6 18H18',
      'M8 21H16'
    ]
  },
  delete: {
    paths: [
      'M5 7H19',
      'M10 11V17',
      'M14 11V17',
      'M8 7L9 4H15L16 7',
      'M7 7L8 20H16L17 7'
    ]
  },
  distribution: {
    paths: [
      'M4 8H15',
      'M4 12H20',
      'M4 16H11',
      'M16 8H20',
      'M12 16H20'
    ]
  },
  edit: {
    paths: [
      'M4 20H9',
      'M14.5 5.5L18.5 9.5L9 19H5V15L14.5 5.5Z',
      'M13 7L17 11'
    ]
  },
  layer: {
    paths: [
      'M12 3L21 8L12 13L3 8L12 3Z',
      'M5 12L12 16L19 12',
      'M5 16L12 20L19 16'
    ]
  },
  lineChart: {
    paths: [
      'M4 19H20',
      'M4 19V5',
      'M7 15L10 11L13 13L18 7',
      'M7 15H7.01',
      'M10 11H10.01',
      'M13 13H13.01',
      'M18 7H18.01'
    ]
  },
  rank: {
    paths: [
      'M8 21V10',
      'M16 21V6',
      'M12 21V14',
      'M5 21H19',
      'M16 3L17 5L19 5.3L17.5 6.8L17.9 9L16 8L14.1 9L14.5 6.8L13 5.3L15 5L16 3Z'
    ]
  },
  scores: {
    paths: [
      'M6 3H18V21H6V3Z',
      'M9 7H15',
      'M9 11H15',
      'M9 15H12',
      'M15 16L17 18L21 14'
    ]
  },
  student: {
    paths: [
      'M12 12C14.2 12 16 10.2 16 8C16 5.8 14.2 4 12 4C9.8 4 8 5.8 8 8C8 10.2 9.8 12 12 12Z',
      'M5 21C5.8 16.8 8.4 15 12 15C15.6 15 18.2 16.8 19 21'
    ]
  },
  sync: {
    paths: [
      'M20 7H10C7.8 7 6 8.8 6 11',
      'M17 4L20 7L17 10',
      'M4 17H14C16.2 17 18 15.2 18 13',
      'M7 20L4 17L7 14'
    ]
  },
  teacher: {
    paths: [
      'M12 11C14 11 15.5 9.5 15.5 7.5C15.5 5.5 14 4 12 4C10 4 8.5 5.5 8.5 7.5C8.5 9.5 10 11 12 11Z',
      'M5 20C5.7 16.2 8.1 14.5 12 14.5C15.9 14.5 18.3 16.2 19 20',
      'M17 6H21V16H18.5'
    ]
  },
  trend: {
    paths: [
      'M4 17L9 12L13 15L20 7',
      'M15 7H20V12',
      'M4 21H20'
    ]
  },
  upload: {
    paths: [
      'M12 16V4',
      'M7 9L12 4L17 9',
      'M5 20H19',
      'M6 16V20',
      'M18 16V20'
    ]
  },
  warning: {
    paths: [
      'M12 4L22 20H2L12 4Z',
      'M12 10V14',
      'M12 18H12.01'
    ]
  }
}
