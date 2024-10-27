import {Box, Button, Input} from '@mui/joy'
import type {ChangeEvent} from 'react'
import {useState} from 'react'
import {skipToken, useQuery} from '@tanstack/react-query'
import {useNavigate} from '@tanstack/react-router'

interface LoginResponse {
	User: {
		username: string
		fullName: string
		email: string
		passwordChangedAt: string
		createdAt: string
		updatedAt: string
	}
	AccessToken: ''
}

const getUser = async (username: string, password: string) => {
	try {
		const res = await fetch('http://localhost:30001/users/login', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({username, password}),
		})
		const data: LoginResponse = await res.json()
		return data
	} catch (error) {
		if (error instanceof Error) {
			throw new Error(error.message)
		}
		throw new Error(String(error))
	}
}

/**
 * @returns JSXElement
 */
export default function Login() {
	const [username, setUsername] = useState<string>('')
	const [password, setPassword] = useState<string>('')
	const [query, setQuery] = useState<boolean>(false)

	const {isError, data, error} = useQuery({
		queryKey: ['login', username, password],
		queryFn: query ? () => getUser(username, password) : skipToken,
	})

	if (isError) {
		return <span>Error: {error.message}</span>
	}

	const handleLogin = () => {
		setQuery(true)
	}

	const navigate = useNavigate({from: '/login'})
	if (data) {
		console.log('data', data)
		navigate({
			to: '/',
		})
	}

	return (
		<Box>
			<Input
				placeholder="Username"
				variant="soft"
				color="primary"
				required
				sx={{
					'--Input-focusedInset': 'var(--any, )',
					'--Input-focusedThickness': '0.12rem',
					'--Input-focusedHighlight': 'rgba(13,110,253,.25)',
					'&::before': {
						transition: 'box-shadow .15s ease-in-out',
					},
					'&:focus-within': {
						borderColor: '#86b7fe',
					},
				}}
				onChange={(e: ChangeEvent<HTMLInputElement>) =>
					setUsername(e.target.value)
				}
			/>
			<Input
				placeholder="Password"
				variant="soft"
				color="primary"
				required
				sx={{
					'--Input-focusedInset': 'var(--any, )',
					'--Input-focusedThickness': '0.12rem',
					'--Input-focusedHighlight': 'rgba(13,110,253,.25)',
					'&::before': {
						transition: 'box-shadow .15s ease-in-out',
					},
					'&:focus-within': {
						borderColor: '#86b7fe',
					},
				}}
				onChange={(e: ChangeEvent<HTMLInputElement>) =>
					setPassword(e.target.value)
				}
			/>
			<Button
				type="button"
				onClick={() => handleLogin()}
			>
				Submit
			</Button>
		</Box>
	)
}
