import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';

/**
 * Centralised invalidation hook for all company-scoped React Query caches.
 * Call this after any mutation (create, edit, delete, bulk op) so the UI
 * reflects the new data immediately instead of waiting for staleTime.
 *
 * Add new query keys here as the app grows — one place, consistent behaviour.
 */
export function useInvalidateCompany(companyId: string) {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['shipments', companyId] });
    queryClient.invalidateQueries({ queryKey: ['shipments-list', companyId] });
    queryClient.invalidateQueries({ queryKey: ['company', companyId] });
  }, [companyId, queryClient]);
}
