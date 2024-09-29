import {createLazyFileRoute} from '@tanstack/react-router'

export const Route = createLazyFileRoute('/')({
	component: Index,
})

/**
 * @returns JSXElement
 */
export default function Index() {
	return <> Hei</>
}
