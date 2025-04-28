import { NavigateFunction } from 'react-router-dom';

let navigate: NavigateFunction | null = null;

export const setNavigate = (navigateFn: NavigateFunction) => {
    navigate = navigateFn;
};

export const navigationService = {
    toAuth: () => {
        if (navigate) {
            navigate('/auth');
        } else {
            console.warn('Navigation not initialized. Falling back to window.location');
            window.location.href = '/auth';
        }
    },
    toHome: () => {
        if (navigate) {
            navigate('/');
        } else {
            console.warn('Navigation not initialized. Falling back to window.location');
            window.location.href = '/';
        }
    }
}; 