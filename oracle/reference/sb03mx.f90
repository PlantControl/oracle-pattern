program sb03mx_probes
  implicit none
  integer :: n,i,j,k,kase,info,state(3),iteration
  integer, allocatable :: signs(:)
  real(8), allocatable :: a(:,:),x(:,:),v(:),work(:)
  real(8) :: est,scale
  character :: trans,op
  read(*,*) n,trans
  allocate(a(n,n),x(n,n),v(n*n),work(2*n),signs(n*n))
  do i=1,n
    read(*,*) (a(i,j),j=1,n)
  end do
  est=0; kase=0; iteration=0
  do
    call dlacn2(n*n,v,x,signs,est,kase,state)
    if(kase==0) exit
    iteration=iteration+1
    op=trans
    if(kase/=1) then
      if(trans=='N') then
        op='T'
      else
        op='N'
      end if
    end if
    write(*,*) iteration,kase,op
    write(*,'(*(ES26.17E3,1X))') ((x(i,j),j=1,n),i=1,n)
    call sb03mx(op,n,a,n,x,n,scale,work,info)
    write(*,*) info
    write(*,'(*(ES26.17E3,1X))') scale
    write(*,'(*(ES26.17E3,1X))') ((x(i,j),j=1,n),i=1,n)
  end do
end program
