import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { endpoints } from './api'

export function useMe() {
  return useQuery({ queryKey: ['me'], queryFn: () => endpoints.me() })
}

export function useClients() {
  return useQuery({ queryKey: ['clients'], queryFn: () => endpoints.clients() })
}

export function useSchedule(from: string, to: string) {
  return useQuery({
    queryKey: ['lessons', from, to],
    queryFn: () => endpoints.lessons(from, to),
    enabled: !!from && !!to,
  })
}

export function useAttendance(lessonId: number) {
  return useQuery({
    queryKey: ['attendance', lessonId],
    queryFn: () => endpoints.attendance(lessonId),
    enabled: !!lessonId,
  })
}

export function useMarkAttendance(lessonId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { clientId: number; present: boolean }) =>
      endpoints.markAttendance(lessonId, vars.clientId, vars.present),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['attendance', lessonId] }),
  })
}

export function useDebtors(days: number) {
  return useQuery({
    queryKey: ['debtors', days],
    queryFn: () => endpoints.debtors(days),
    enabled: !!days,
  })
}

export function useWaitingList() {
  return useQuery({ queryKey: ['waiting'], queryFn: () => endpoints.waitingList() })
}

export function useAddWaitingList() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (clientId: number) => endpoints.addWaitingList(clientId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waiting'] }),
  })
}

export function useRemoveWaitingList() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => endpoints.removeWaitingList(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waiting'] }),
  })
}

export function useSubmitWellbeing() {
  return useMutation({
    mutationFn: (vars: { lessonId: number; wellbeing: number; note: string }) =>
      endpoints.wellbeing(vars.lessonId, vars.wellbeing, vars.note),
  })
}

export function useWellbeingHistory(clientId: number) {
  return useQuery({
    queryKey: ['wellbeing', clientId],
    queryFn: () => endpoints.wellbeingHistory(clientId),
    enabled: !!clientId,
  })
}

export function useMessageCoach() {
  return useMutation({
    mutationFn: (vars: { from: string; text: string }) => endpoints.messageCoach(vars.from, vars.text),
  })
}

export function useSocialMedia() {
  return useQuery({ queryKey: ['social'], queryFn: () => endpoints.socialMedia() })
}

export function useFaq(q: string) {
  return useQuery({
    queryKey: ['faq', q],
    queryFn: () => endpoints.faq(q),
    enabled: q.trim().length > 0,
  })
}
