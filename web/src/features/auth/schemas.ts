import { z } from 'zod'

export const phoneSchema = z
  .string()
  .min(1, 'Введите телефон')
  .max(12, 'Номер должен быть 11 цифр')
  .refine((v) => /^(\+7|8)\d{10}$/.test(v.trim().replace(/\s|-|\(|\)/g, '')), 'Некорректный номер. Формат: +7 XXX XXX XX XX')

export const passwordSchema = z
  .string()
  .min(6, 'Пароль минимум 6 символов')

export const loginSchema = z.object({
  phone: phoneSchema,
  password: z.string().min(1, 'Введите пароль'),
})

export const registerSchema = z.object({
  firstName: z.string().min(1, 'Введите имя').max(64),
  lastName: z.string().max(64).optional().or(z.literal('')),
  phone: phoneSchema,
  password: passwordSchema,
})

export type LoginValues = z.infer<typeof loginSchema>
export type RegisterValues = z.infer<typeof registerSchema>
