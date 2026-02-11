import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import BtTablePassword from './index.vue'

const mockCopyText = vi.fn()

vi.mock('@/hooks/useCopy', () => ({
  useCopy: () => ({
    copyText: mockCopyText,
  }),
}))

// Stub naive-ui ellipsis
const NEllipsis = {
  template: '<span><slot /></span>',
  props: ['tooltip'],
}

describe('BtTablePassword', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mountComponent = (props = {}) => {
    return mount(BtTablePassword, {
      props: { value: 'secret123', ...props },
      global: {
        stubs: {
          'n-ellipsis': NEllipsis,
        },
      },
    })
  }

  it('renders masked password by default', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('**********')
    expect(wrapper.text()).not.toContain('secret123')
  })

  it('shows password after toggle click', async () => {
    const wrapper = mountComponent()
    // Click the eye toggle
    const toggleBtn = wrapper.find('[title="Show"]')
    await toggleBtn.trigger('click')
    expect(wrapper.text()).toContain('secret123')
  })

  it('hides password again on second toggle', async () => {
    const wrapper = mountComponent()
    const toggleBtn = wrapper.find('[title="Show"]')
    await toggleBtn.trigger('click')
    await toggleBtn.trigger('click')
    expect(wrapper.text()).toContain('**********')
  })

  it('calls copyText with the value on copy click', async () => {
    const wrapper = mountComponent({ value: 'mypass' })
    const copyBtn = wrapper.find('[title="Copy"]')
    await copyBtn.trigger('click')
    expect(mockCopyText).toHaveBeenCalledWith('mypass')
  })

  it('handles number value prop', async () => {
    const wrapper = mountComponent({ value: 12345 })
    const copyBtn = wrapper.find('[title="Copy"]')
    await copyBtn.trigger('click')
    expect(mockCopyText).toHaveBeenCalledWith('12345')
  })
})
