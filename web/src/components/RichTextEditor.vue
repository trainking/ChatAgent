<template>
  <div class="rte-wrapper">
    <div class="rte-toolbar">
      <button
        type="button"
        class="rte-btn"
        title="Bold"
        @click="exec('bold')"
      >
        <strong>B</strong>
      </button>
      <button
        type="button"
        class="rte-btn"
        title="Italic"
        @click="exec('italic')"
      >
        <em>I</em>
      </button>
      <button
        type="button"
        class="rte-btn"
        title="Underline"
        @click="exec('underline')"
      >
        <u>U</u>
      </button>
    </div>
    <div
      ref="editorRef"
      class="rte-content"
      contenteditable="true"
      @input="onInput"
      @paste="onPaste"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'

const props = defineProps<{ modelValue: string; maxlength?: number }>()
const emit = defineEmits<{ (e: 'update:modelValue', val: string): void }>()

const editorRef = ref<HTMLDivElement>()

function exec(command: string) {
  document.execCommand(command, false)
  editorRef.value?.focus()
  emitInput()
}

function emitInput() {
  if (!editorRef.value) return
  let html = editorRef.value.innerHTML
  if (props.maxlength && html.length > props.maxlength) {
    html = html.substring(0, props.maxlength)
  }
  emit('update:modelValue', html)
}

function onInput() {
  emitInput()
}

function onPaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text/plain') || ''
  document.execCommand('insertText', false, text)
}

watch(() => props.modelValue, (val) => {
  if (editorRef.value && editorRef.value.innerHTML !== val) {
    editorRef.value.innerHTML = val || ''
  }
})

onMounted(() => {
  nextTick(() => {
    if (editorRef.value) {
      editorRef.value.innerHTML = props.modelValue || ''
    }
  })
})
</script>

<style scoped>
.rte-wrapper {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}
.rte-toolbar {
  display: flex;
  gap: 2px;
  padding: 4px 6px;
  background: #f5f7fa;
  border-bottom: 1px solid #dcdfe6;
}
.rte-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 3px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #606266;
}
.rte-btn:hover {
  background: #e6e8eb;
  color: #409eff;
}
.rte-content {
  min-height: 80px;
  padding: 8px 12px;
  outline: none;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}
.rte-content:empty::before {
  content: attr(data-placeholder);
  color: #c0c4cc;
}
</style>
