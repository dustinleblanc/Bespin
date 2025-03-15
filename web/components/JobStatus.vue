<template>
  <div :class="['status', `status-${status}`]" @click="$emit('status-click', status)">
    {{ status }}
  </div>
</template>

<script setup lang="ts">
defineProps<{
  status: 'pending' | 'running' | 'completed' | 'failed'
}>()

defineEmits<{
  'status-click': [status: string]
}>()
</script>

<style scoped>
.status {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
}

.status-pending {
  background-color: #f3f4f6;
  color: #374151;
}

.status-running {
  background-color: #dbeafe;
  color: #1e40af;
}

.status-completed {
  background-color: #dcfce7;
  color: #166534;
}

.status-failed {
  background-color: #fee2e2;
  color: #991b1b;
}
</style>
