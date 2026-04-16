<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { tagApi } from '@/api'
import { toast } from 'vue-sonner'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import { Plus, X } from 'lucide-vue-next'

interface Tag {
  ID: string
  name: string
  color: string
}

const tags = ref<Tag[]>([])
const loading = ref(true)
const showCreateDialog = ref(false)

const PRESET_COLORS = [
  '#3498db', '#e74c3c', '#2ecc71', '#f39c12', '#9b59b6',
  '#1abc9c', '#e67e22', '#34495e', '#16a085', '#c0392b',
]

const newTag = ref({ name: '', color: '#3498db' })
const creating = ref(false)

async function fetchTags() {
  try {
    const res: any = await tagApi.list()
    if (res.code === 0) {
      tags.value = res.data || []
    }
  } catch (e) {
    console.error('Failed to load tags:', e)
  } finally {
    loading.value = false
  }
}

async function createTag() {
  const name = newTag.value.name.trim()
  if (!name) {
    toast.error('请输入标签名称')
    return
  }
  creating.value = true
  try {
    const res: any = await tagApi.create({ name, color: newTag.value.color })
    if (res.code === 0) {
      tags.value.push(res.data)
      newTag.value = { name: '', color: '#3498db' }
      showCreateDialog.value = false
      toast.success('标签创建成功')
    } else {
      toast.error(res.msg || '创建失败')
    }
  } catch (e) {
    toast.error('创建请求失败')
  } finally {
    creating.value = false
  }
}

async function deleteTag(tag: Tag) {
  if (!confirm(`确定删除标签「${tag.name}」？关联的账单不会被删除。`)) return
  try {
    const res: any = await tagApi.remove(tag.ID)
    if (res.code === 0) {
      tags.value = tags.value.filter(t => t.ID !== tag.ID)
      toast.success('已删除')
    } else {
      toast.error(res.msg || '删除失败')
    }
  } catch (e) {
    toast.error('删除请求失败')
  }
}

onMounted(() => {
  fetchTags()
})

onUnmounted(() => {
  toast.dismiss()
})
</script>

<template>
  <div class="tag-page">
    <TopNavBar title="标签管理" />

    <div class="tag-content">
      <!-- 标签列表 -->
      <div v-if="loading" class="loading-hint">加载中...</div>

      <div v-else-if="tags.length === 0" class="empty-hint">
        <p>还没有创建任何标签</p>
        <p class="sub">标签可以帮助你灵活分类账单，如 #出差、#报销、#AA</p>
      </div>

      <div v-else class="tag-list">
        <div v-for="tag in tags" :key="tag.ID" class="tag-item card">
          <div class="tag-left">
            <span class="tag-dot" :style="{ background: tag.color }"></span>
            <span class="tag-name">{{ tag.name }}</span>
          </div>
          <button class="tag-delete-btn" @click="deleteTag(tag)">
            <X :size="16" />
          </button>
        </div>
      </div>

      <!-- 新建按钮 -->
      <button class="fab-btn" @click="showCreateDialog = true">
        <Plus :size="24" />
      </button>

      <!-- 新建弹窗 -->
      <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
        <div class="dialog card">
          <h3>新建标签</h3>
          <div class="form-group">
            <label>标签名称</label>
            <input
              type="text"
              v-model="newTag.name"
              class="form-control"
              placeholder="例如：出差、报销、AA"
              maxlength="50"
              @keyup.enter="createTag"
            />
          </div>
          <div class="form-group">
            <label>标签颜色</label>
            <div class="color-picker">
              <button
                v-for="c in PRESET_COLORS" :key="c"
                :class="['color-dot', { active: newTag.color === c }]"
                :style="{ background: c }"
                @click="newTag.color = c"
              ></button>
            </div>
          </div>
          <div class="dialog-actions">
            <button class="btn btn-ghost" @click="showCreateDialog = false">取消</button>
            <button class="btn btn-primary" @click="createTag" :disabled="creating">
              {{ creating ? '创建中...' : '创建' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tag-page {
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
  max-width: 600px;
  margin: 0 auto;
  min-height: 100vh;
  background: var(--bg-body);
}

.tag-content {
  position: relative;
  min-height: 300px;
}

.loading-hint, .empty-hint {
  text-align: center;
  color: var(--text-secondary);
  padding: 60px 20px;
}
.empty-hint p { margin: 0 0 4px 0; font-weight: 600; }
.empty-hint .sub { font-size: 0.85rem; font-weight: 400; opacity: 0.7; }

.tag-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tag-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  transition: all 0.15s;
}
.tag-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tag-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  flex-shrink: 0;
}
.tag-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
}
.tag-delete-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
}
.tag-delete-btn:active {
  background: rgba(231, 76, 60, 0.1);
  color: #e74c3c;
}

/* FAB */
.fab-btn {
  position: fixed;
  bottom: calc(100px + env(safe-area-inset-bottom));
  right: 24px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 20px rgba(26, 188, 156, 0.4);
  cursor: pointer;
  transition: all 0.2s;
}
.fab-btn:active {
  transform: scale(0.92);
}

/* Dialog */
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.dialog {
  width: 100%;
  max-width: 400px;
  padding: 24px;
  border-radius: 20px;
  animation: dialogIn 0.25s ease-out;
}
@keyframes dialogIn {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
.dialog h3 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0 0 20px 0;
  color: var(--text-primary);
}

.form-group {
  margin-bottom: 16px;
}
.form-group label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 600;
  margin-bottom: 8px;
  display: block;
}
.form-control {
  width: 100%;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  padding: 12px 16px;
  font-size: 1rem;
  font-family: inherit;
  background: #fafafa;
  outline: none;
  color: var(--text-primary);
  transition: all 0.2s;
}
.form-control:focus {
  border-color: var(--primary);
  background: white;
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.1);
}

.color-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.color-dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 3px solid transparent;
  cursor: pointer;
  transition: all 0.2s;
}
.color-dot.active {
  border-color: var(--text-primary);
  transform: scale(1.15);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.color-dot:active {
  transform: scale(0.95);
}

.dialog-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}
.btn {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  font-size: 0.95rem;
  font-weight: 700;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  text-align: center;
}
.btn-primary {
  background: var(--primary);
  color: white;
}
.btn-primary:disabled { opacity: 0.6; }
.btn-ghost {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-secondary);
}
</style>
