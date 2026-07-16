import { z } from 'zod'

export const phoneSchema = z
  .string()
  .min(1, 'Введите телефон')
  .refine((v) => /^\+?\d[\d\s()-]{9,}$/.test(v.trim()), 'Некорректный номер телефона')

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
