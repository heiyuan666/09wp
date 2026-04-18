export function usePagination<T>(items: Ref<T[]>, itemsPerPage = 12) {
  const currentPage = ref(1)
  
  const totalPages = computed(() => Math.ceil(items.value.length / itemsPerPage))
  
  const paginatedItems = computed(() => {
    const start = (currentPage.value - 1) * itemsPerPage
    const end = start + itemsPerPage
    return items.value.slice(start, end)
  })
  
  const hasNextPage = computed(() => currentPage.value < totalPages.value)
  const hasPrevPage = computed(() => currentPage.value > 1)
  
  function goToPage(page: number) {
    if (page >= 1 && page <= totalPages.value) {
      currentPage.value = page
    }
  }
  
  function nextPage() {
    if (hasNextPage.value) {
      currentPage.value++
    }
  }
  
  function prevPage() {
    if (hasPrevPage.value) {
      currentPage.value--
    }
  }
  
  function resetPage() {
    currentPage.value = 1
  }
  
  // Reset page when items change
  watch(items, () => {
    if (currentPage.value > totalPages.value) {
      currentPage.value = Math.max(1, totalPages.value)
    }
  })
  
  return {
    currentPage,
    totalPages,
    paginatedItems,
    hasNextPage,
    hasPrevPage,
    goToPage,
    nextPage,
    prevPage,
    resetPage,
    totalItems: computed(() => items.value.length),
  }
}
