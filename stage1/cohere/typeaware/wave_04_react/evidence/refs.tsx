import {useRef, useState, useCallback, useMemo, useReducer, useEffect} from 'react';
// @validateRefAccessDuringRender
function Component(props) {
  const ref = useRef(null);
  const value = ref.current;
  return value;
}
